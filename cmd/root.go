// TODO clean this mess of a file
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/truecrunchyfrog/ian"
)

var cfgFile string

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
	viper.BindPFlag("root", rootCmd.PersistentFlags().Lookup("root"))
	viper.BindPFlag("timezone", rootCmd.PersistentFlags().Lookup("timezone"))

	rootCmd.Flags().Bool("sunday", false, "Use sunday instead of monday as the first day of the week.")
	rootCmd.Flags().BoolP("weeks", "w", false, "Show week numbers.")
	rootCmd.Flags().BoolVarP(&emptyCalendar, "empty", "e", false, "Do not show any calendar events, just an empty calendar.")
	viper.BindPFlag("sunday", rootCmd.Flags().Lookup("sunday"))
	viper.BindPFlag("weeks", rootCmd.Flags().Lookup("weeks"))
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

	// Cannot rely on rootCmd.MarkFlagsMutuallyExclusive("sunday", "weeks") because it does not work with viper
	if sunday && showWeeks {
		log.Fatal("sundays and week numbers cannot be combined")
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

	instance, err := ian.CreateInstance(GetRoot(), !emptyCalendar)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ian.DisplayCalendar(GetTimeZone(), year, time.Month(month), now, sunday, showWeeks, func(monthDay int, isToday bool) string {
		if isToday {
			return "\033[44m"
		}
		if !emptyCalendar {
			dayStart := time.Date(year, time.Month(month), monthDay, 0, 0, 0, 0, GetTimeZone())
			dayEnd := dayStart.AddDate(0, 0, 1).Add(-time.Second)
			eventsInDay := instance.FilterEvents(func(e ian.Event) bool {
				return ian.DoPeriodsMeet(e.Props.Start, e.Props.End, dayStart, dayEnd)
			})
			switch {
			case len(eventsInDay) > 1:
				return "\033[31m"
			case len(eventsInDay) != 0:
				return "\033[33m"
			}
		}
		return ""
	}))

	// TODO move event list into display.go
	if !emptyCalendar {
		monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, GetTimeZone())
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

		events := instance.FilterEvents(func(e ian.Event) bool {
			return ian.DoPeriodsMeet(e.Props.Start, e.Props.End, monthStart, monthEnd)
		})

    fmt.Println(ian.DisplayTimeline(instance, monthStart, monthEnd, events, GetTimeZone()))
	}
}
