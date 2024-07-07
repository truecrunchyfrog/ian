package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/truecrunchyfrog/ian"
)

var cfgFile string

var viewDate string
var emptyCalendar bool

var rootCmd = &cobra.Command{
	Use:   "ian",
	Short: "ian is a file-based calendar and reminder system.",
	Run:   rootCmdRun,
}

func Execute() error {
	return rootCmd.Execute()
}

func GetRoot() string {
	dir := viper.GetString("root")
	ian.CreateDir(dir)
	return dir
}

func GetTimeZone() time.Location {
	timeZoneFlag := viper.GetString("timezone")

	if timeZoneFlag == "" {
		return *time.Local
	}

	t1, err := time.Parse("MST", timeZoneFlag)

	if err == nil {
		return *t1.Location()
	} else {
		t2, err := time.Parse("-0700", timeZoneFlag)
		if err != nil {
			log.Fatal("invalid time zone '" + timeZoneFlag + "'")
		}

		return *t2.Location()
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
	rootCmd.PersistentFlags().String("timezone", "", "Override the automatic local timezone.")
	rootCmd.PersistentFlags().BoolVarP(&ian.Verbose, "verbose", "v", false, "Enable verbose mode. More information is given.")
	viper.BindPFlag("root", rootCmd.PersistentFlags().Lookup("root"))
	viper.BindPFlag("timezone", rootCmd.PersistentFlags().Lookup("timezone"))

	rootCmd.Flags().StringVarP(&viewDate, "month", "m", "", "The month to view.")
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

	now := time.Now()

	var t time.Time
	if viewDate != "" {
		var err error
		if t, err = ian.ParseYearAndMonth(viewDate); err != nil {
			log.Fatal(err)
		}
	} else {
		t = now // Default to current date
	}

	instance, err := ian.CreateInstance(GetRoot(), !emptyCalendar)
	if err != nil {
		log.Fatal(err)
	}

	tz := GetTimeZone()
	fmt.Println(ian.DisplayCalendar(t, now, sunday, showWeeks, func(monthDay int, isToday bool) string {
		if isToday {
			return "\033[44m"
		}
		if !emptyCalendar {
			day := time.Date(t.Year(), t.Month(), monthDay, 0, 0, 0, 0, &tz)
			dayAfter := time.Date(t.Year(), t.Month(), monthDay+1, 0, 0, 0, 0, &tz)
			eventsStartingHere := instance.FilterEvents(func(e ian.Event) bool {
				return !e.Properties.Start.Before(day) && e.Properties.Start.Before(dayAfter)
			})
			switch {
			case len(eventsStartingHere) > 1:
				return "\033[31m"
			case len(eventsStartingHere) != 0:
				return "\033[33m"
			}
		}
		return ""
	}))

  if !emptyCalendar {
  }
}
