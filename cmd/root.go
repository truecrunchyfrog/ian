package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/acarl005/stripansi"
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

func checkCollision(events *[]ian.Event, props ian.EventProperties) {
	if !ignoreCollisionWarnings || noCollision {
		collidingEvents := ian.FilterEvents(events, func(e *ian.Event) bool {
			return e.Props.Uid != props.Uid && !slices.Contains(collisionExceptions, e.Path.Calendar()) && ian.DoPeriodsMeet(props.GetTimeRange(), e.Props.GetTimeRange())
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
	rootCmd.PersistentFlags().Bool("no-validation", false, "Disable event validation. Used for updating invalid events.")
	rootCmd.PersistentFlags().BoolVarP(&ian.Verbose, "verbose", "v", false, "Enable verbose mode. More information is given.")
	rootCmd.PersistentFlags().BoolVar(&noCollision, "no-collision", false, "Prevent events from being created or edited to collide with another event.")
	rootCmd.PersistentFlags().StringSliceVar(&collisionExceptions, "collision-exceptions", []string{}, "Mark a list of `calendars` as exceptions for collisions. When a calendar is listed, collision warnings will not be shown, and when combined with 'no-collision' a collision for an event within a calendar specified here will pass.")
	rootCmd.PersistentFlags().BoolVar(&ignoreCollisionWarnings, "no-collision-warnings", false, "Hides the warnings shown when an event will collide with an existing event.")
	viper.BindPFlag("root", rootCmd.PersistentFlags().Lookup("root"))
	viper.BindPFlag("timezone", rootCmd.PersistentFlags().Lookup("timezone"))
	viper.BindPFlag("no-validation", rootCmd.PersistentFlags().Lookup("no-validation"))
	viper.BindPFlag("no-collision", rootCmd.PersistentFlags().Lookup("no-collision"))
	viper.BindPFlag("collision-exceptions", rootCmd.PersistentFlags().Lookup("collision-exceptions"))
	viper.BindPFlag("no-collision-warnings", rootCmd.PersistentFlags().Lookup("no-collision-warnings"))

	rootCmd.Flags().Int("first-weekday", 2, "Specify the first day of the week by index (1 = Sunday, 2 = Monday, ... 7 = Saturday).")
	rootCmd.Flags().BoolP("weeks", "w", false, "Show week numbers.")
  rootCmd.Flags().IntP("months", "m", 3, "Total months to show. If more than one, the rest will be the months following the first.")
	rootCmd.Flags().BoolVarP(&emptyCalendar, "empty", "e", false, "Just an empty calendar.")
	rootCmd.Flags().Bool("no-timeline", false, "Do not show the timeline.")
	rootCmd.Flags().Bool("no-event-coloring", false, "Do not color the calendar days based on events occuring then.")
	rootCmd.Flags().Bool("no-day-hinting", false, "Do not color the current day.")
	rootCmd.Flags().Uint("daywidth", 3, "Width per calendar day in character length.")
	rootCmd.Flags().Bool("no-legend", false, "Do not show the calendar legend that shows what colors belong to what calendar.")
	viper.BindPFlag("first-weekday", rootCmd.Flags().Lookup("first-weekday"))
	viper.BindPFlag("weeks", rootCmd.Flags().Lookup("weeks"))
	viper.BindPFlag("months", rootCmd.Flags().Lookup("months"))
	viper.BindPFlag("no-timeline", rootCmd.Flags().Lookup("no-timeline"))
	viper.BindPFlag("no-event-coloring", rootCmd.Flags().Lookup("no-event-coloring"))
	viper.BindPFlag("daywidth", rootCmd.Flags().Lookup("daywidth"))
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
	firstWeekday := time.Weekday(viper.GetInt("first-weekday") - 1)
	showWeeks := viper.GetBool("weeks")
	widthPerDay := viper.GetUint("daywidth")
  months := viper.GetInt("months")

	if firstWeekday != time.Monday && showWeeks {
		log.Fatal("week numbers only if the first weekday is monday")
	}

	if widthPerDay < 2 || widthPerDay > 100 {
		log.Fatal("invalid daywidth size (must be within the bounds of 2-100)")
	}

	now := time.Now().In(ian.GetTimeZone())

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

	var events []ian.Event

	if !emptyCalendar {
		events, _, err = instance.ReadEvents(ian.TimeRange{})
		if err != nil {
			log.Fatal(err)
		}
	}

	var leftSide, rightSide []string

	leftSide = strings.Split(ian.DisplayCalendar(
		ian.GetTimeZone(),
		year,
		time.Month(month),
    months,
		firstWeekday,
		showWeeks,
		int(widthPerDay),
		func(y int, m time.Month, d int) (string, bool) {
			if y == now.Year() && m == now.Month() && d == now.Day() {
				return "\033[1;30;47m", true
			}
			if !emptyCalendar && !viper.GetBool("no-event-coloring") {
				var dayRange ian.TimeRange
				dayRange.From = time.Date(y, m, d, 0, 0, 0, 0, ian.GetTimeZone())
				dayRange.To = dayRange.From.AddDate(0, 0, 1)
				eventsInDay := []*ian.Event{}
				for _, event := range events {
					if ian.DoPeriodsMeet(event.Props.GetTimeRange(), dayRange) {
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
							calendar = event.Path.Calendar()
							continue
						}

						if calendar != event.Path.Calendar() {
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
		}), "\n")

	if !emptyCalendar && !viper.GetBool("no-timeline") {
		rightSide = strings.Split(ian.DisplayTimeline(instance, events, ian.GetTimeZone()), "\n")
	}

	if !emptyCalendar && !viper.GetBool("no-legend") {
		leftSide = append(leftSide, strings.Split(ian.DisplayCalendarLegend(instance, events), "\n")...)
	}

	var indent int

	for _, line := range leftSide {
    if l := utf8.RuneCountInString(stripansi.Strip(line)); l > indent {
			indent = l + 5
		}
	}

	for i := 0; i < max(len(leftSide), len(rightSide)); i++ {
		var line string

		if i < len(leftSide) {
			line += leftSide[i]
		}

		if i < len(rightSide) {
			line += strings.Repeat(" ", indent-utf8.RuneCountInString(stripansi.Strip(line)))
			line += rightSide[i]
		}

		fmt.Println(line)
	}
}
