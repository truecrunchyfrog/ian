package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/truecrunchyfrog/ian"
)

var cfgFile string

var viewDate string
var firstWeekday int

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
	rootCmd.Flags().IntVarP(&firstWeekday, "firstweekday", "w", 0, "The first day of the week, for display in the calendar.")
	viper.BindPFlag("firstweekday", rootCmd.Flags().Lookup("firstweekday"))
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
	if firstWeekday < 0 || firstWeekday > 6 {
		log.Fatal("'firstweekday' out of bounds 0-6")
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

	fmt.Println(ian.DisplayCalendar(t, now, time.Weekday(firstWeekday), true, func(monthDay int, isToday bool) string {
    r := fmt.Sprint(monthDay)
    if isToday {
      r = "\033[44m" + r
    }
    return r
  }))
}
