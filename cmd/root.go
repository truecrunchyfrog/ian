package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/truecrunchyfrog/ian"
)

var cfgFile string
var noCollision bool
var collisionExceptions []string
var ignoreCollisionWarnings bool

var emptyCalendar bool

func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "ian [month] [year]",
	Short: "ian is a file-based calendar",
	Args:  cobra.RangeArgs(0, 2),
	Run:   rootCmdRun,
}

func GetRoot() string {
	dir := viper.GetString("root")
	ian.CreateDir(dir)
	return dir
}

var TimeZone *time.Location

func GetTimeZone() *time.Location {
	if TimeZone != nil {
		return TimeZone
	}

	timeZoneFlag := viper.GetString("timezone")

	if timeZoneFlag == "" {
		return time.Local
	}

	t1, err := time.Parse("MST", timeZoneFlag)

	if err == nil {
		TimeZone = t1.Location()
	} else {
		t2, err := time.Parse("-0700", timeZoneFlag)
		if err != nil {
			log.Fatal("invalid time zone '" + timeZoneFlag + "'")
		}

		TimeZone = t2.Location()
	}

	return TimeZone
}

func checkCollision(events *[]ian.Event, props ian.EventProperties) {
	if !ignoreCollisionWarnings || noCollision {
		collidingEvents := ian.FilterEvents(events, func(e *ian.Event) bool {
			return !slices.Contains(collisionExceptions, e.GetCalendarName()) && ian.DoPeriodsMeet(props.Start, props.End, e.Props.Start, e.Props.End)
		})

		if !ignoreCollisionWarnings {
			for _, collidingEvent := range collidingEvents {
				fmt.Printf("warning: this event collides with '%s'.\n", collidingEvent.Props.Summary)
			}
		}

		if len(collidingEvents) != 0 && noCollision {
			log.Fatal("COLLISION! cannot set event in that period, because it collides with other events.\nif you want to set it anyway, either disable the 'no-collision' preference/flag or add a '--no-collision=false' flag to temporarily override it.")
		}
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Configuration file. Defaults are '$HOME/.ian.toml', then '$HOME/.config/ian/.ian.toml'.")

	var defaultRoot string
	if home, err := os.UserHomeDir(); err == nil {
		defaultRoot = filepath.Join(home, ".ian")
	} else {
		fmt.Println("warning: home directory environment variable missing")
	}

	rootCmd.PersistentFlags().String("root", defaultRoot, "Set the calendar root.")
	rootCmd.PersistentFlags().String("timezone", "", "Set the timezone to work with. Overrides the local timezone.")
	rootCmd.PersistentFlags().BoolVarP(&ian.Verbose, "verbose", "v", false, "Enable verbose mode. More information is given.")
	rootCmd.PersistentFlags().BoolVar(&noCollision, "no-collision", false, "Prevent events from being created or edited to collide with another event.")
	rootCmd.PersistentFlags().StringSliceVar(&collisionExceptions, "collision-exceptions", []string{}, "Mark a list of `calendars` as exceptions for collisions. When a calendar is listed, collision warnings will not be shown, and when combined with 'no-collision' a collision for an event within a calendar specified here will pass.")
	rootCmd.PersistentFlags().BoolVar(&ignoreCollisionWarnings, "no-collision-warnings", false, "Hides the warnings shown when an event will collide with an existing event.")
	viper.BindPFlag("root", rootCmd.PersistentFlags().Lookup("root"))
	viper.BindPFlag("timezone", rootCmd.PersistentFlags().Lookup("timezone"))
	viper.BindPFlag("no-collision", rootCmd.PersistentFlags().Lookup("no-collision"))
	viper.BindPFlag("collision-exceptions", rootCmd.PersistentFlags().Lookup("collision-exceptions"))
	viper.BindPFlag("no-collision-warnings", rootCmd.PersistentFlags().Lookup("no-collision-warnings"))

	rootCmd.Flags().Bool("sunday", false, "Use sunday instead of monday as the first day of the week.")
	rootCmd.Flags().BoolP("weeks", "w", false, "Show week numbers.")
	rootCmd.Flags().BoolVarP(&emptyCalendar, "empty", "e", false, "Just an empty calendar.")
	rootCmd.Flags().Bool("no-timeline", false, "Do not show the timeline.")
	rootCmd.Flags().Bool("no-event-coloring", false, "Do not color the calendar days based on events occuring then.")
	rootCmd.Flags().Bool("no-day-hinting", false, "Do not color the current day.")
	rootCmd.Flags().Bool("no-week-hinting", false, "Do not show a brighter color on the current week and weekday.")
	rootCmd.Flags().Bool("borders", false, "Show week/weekday borders.")
	rootCmd.Flags().Uint("daywidth", 3, "Width per calendar day in character length.")
	rootCmd.Flags().BoolP("past", "p", false, "Include past events in the timeline.")
	rootCmd.Flags().Bool("no-legend", false, "Do not show the calendar legend that shows what colors belong to what calendar.")
	viper.BindPFlag("sunday", rootCmd.Flags().Lookup("sunday"))
	viper.BindPFlag("weeks", rootCmd.Flags().Lookup("weeks"))
	viper.BindPFlag("no-timeline", rootCmd.Flags().Lookup("no-timeline"))
	viper.BindPFlag("no-event-coloring", rootCmd.Flags().Lookup("no-event-coloring"))
	viper.BindPFlag("no-week-hinting", rootCmd.Flags().Lookup("no-week-hinting"))
	viper.BindPFlag("borders", rootCmd.Flags().Lookup("borders"))
	viper.BindPFlag("daywidth", rootCmd.Flags().Lookup("daywidth"))
	viper.BindPFlag("past", rootCmd.Flags().Lookup("past"))
	viper.BindPFlag("no-legend", rootCmd.Flags().Lookup("no-legend"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(filepath.Join(home, ".config", "ian"))
		viper.SetConfigType("toml")
		viper.SetConfigName(".ian")
	}

	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func rootCmdRun(cmd *cobra.Command, args []string) {
	sunday := viper.GetBool("sunday")
	showWeeks := viper.GetBool("weeks")
	widthPerDay := viper.GetUint("daywidth")

	// Cannot rely on rootCmd.MarkFlagsMutuallyExclusive("sunday", "weeks") because it does not work with viper
	if sunday && showWeeks {
		log.Fatal("sundays and week numbers cannot be combined")
	}

	if widthPerDay < 2 || widthPerDay > 100 {
		log.Fatal("invalid daywidth size (must be within the bounds of 2-100)")
	}

	now := time.Now().In(GetTimeZone())

	year := now.Year()
	month := int(now.Month())

	var err error
	if len(args) >= 1 {
		// Month provided
		if month, err = strconv.Atoi(args[0]); err != nil {
			log.Fatal(err)
		}
	}
	if len(args) >= 2 {
		// Year provided
		if year, err = strconv.Atoi(args[1]); err != nil {
			log.Fatal(err)
		}
	}

	instance, err := ian.CreateInstance(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

	var monthStart time.Time
	if month != int(now.Month()) || viper.GetBool("past") {
		monthStart = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, GetTimeZone())
	} else {
		monthStart = time.Date(year, time.Month(month), now.Day(), 0, 0, 0, 0, GetTimeZone())
	}
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	var events []ian.Event

	if !emptyCalendar {
		events, err = instance.ReadEvents(ian.TimeRange{From: monthStart, To: monthEnd})
    if err != nil {
      log.Fatal(err)
    }
		events = ian.FilterEvents(&events, func(e *ian.Event) bool {
			return ian.DoPeriodsMeet(e.Props.Start, e.Props.End, monthStart, monthEnd)
		})
	}

	fmt.Println(ian.DisplayCalendar(
		GetTimeZone(),
		year,
		time.Month(month),
		now,
		sunday,
		showWeeks,
		!viper.GetBool("no-week-hinting"),
		viper.GetBool("borders"),
		int(widthPerDay),
		func(monthDay int, isToday bool) (string, bool) {
			if isToday {
				return "\033[44m", true
			}
			if !emptyCalendar && !viper.GetBool("no-event-coloring") {
				dayStart := time.Date(year, time.Month(month), monthDay, 0, 0, 0, 0, GetTimeZone())
				dayEnd := dayStart.AddDate(0, 0, 1).Add(-time.Second)
				eventsInDay := []*ian.Event{}
				for _, event := range events {
					if ian.DoPeriodsMeet(event.Props.Start, event.Props.End, dayStart, dayEnd) {
						eventsInDay = append(eventsInDay, &event)
					}
				}
				switch {
				case len(eventsInDay) == 1:
					return ian.GetEventRgbAnsiSeq(eventsInDay[0], instance, false), false
				case len(eventsInDay) > 1:
					sameCalendar := true
					var calendar string
					for _, event := range eventsInDay {
						if calendar == "" {
							calendar = event.GetCalendarName()
							continue
						}

						if calendar != event.GetCalendarName() {
							sameCalendar = false
							break
						}
					}
					if sameCalendar {
						return ian.GetEventRgbAnsiSeq(eventsInDay[0], instance, false) + "\033[4m", false
					}
					return "\033[4m", false
				default:
					return "\033[2m", false
				}
			}
			return "", false
		}))

	if !emptyCalendar && !viper.GetBool("no-timeline") {
		fmt.Println(ian.DisplayTimeline(instance, events, GetTimeZone()))
	}

	if !emptyCalendar && !viper.GetBool("no-legend") {
		fmt.Println("\n" + ian.DisplayCalendarLegend(instance, events))
	}
}
