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

var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "ian",
		Short: "ian is a file-based calendar and reminder system.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("work in progress!")
		},
	}
)

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

	rootCmd.PersistentFlags().StringP("root", "r", defaultRoot, "Set the calendar root.")
	rootCmd.PersistentFlags().StringP("timezone", "t", "", "Override the automatic local timezone.")
	rootCmd.PersistentFlags().BoolVarP(&ian.Verbose, "verbose", "v", false, "Enable verbose mode. More information is given.")
	viper.BindPFlag("root", rootCmd.PersistentFlags().Lookup("root"))
	viper.BindPFlag("timezone", rootCmd.PersistentFlags().Lookup("timezone"))
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
