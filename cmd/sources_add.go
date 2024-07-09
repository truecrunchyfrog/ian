package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

var lifetime string

func init() {
	sourcesAddCmd.Flags().StringVar(&lifetime, "lifetime", "", "Set the duration (e.g. '1h30m') between updates of this source.")

	sourcesCmd.AddCommand(sourcesAddCmd)
}

var sourcesAddCmd = &cobra.Command{
	Use:   "add <name> <type> <source>",
	Short: "Configure a source. The configuration will be formatted.",
	Long:  "If you have comments in your configuration, configure the source manually instead of using this command.",
	Args:  cobra.ExactArgs(3),
	Run:   sourcesAddCmdRun,
}

func sourcesAddCmdRun(cmd *cobra.Command, args []string) {
	if lifetime != "" {
    if _, err := time.ParseDuration(lifetime); err != nil {
      log.Fatal(err)
    }
	}

  name, _type, source := args[0], args[1], args[2]

	config, err := ian.ReadConfig(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

  if config.Sources == nil {
    config.Sources = map[string]ian.CalendarSource{}
  }

  if _, ok := config.Sources[name]; ok {
    log.Fatalf("a source with the name '%s' is already configured.\n", name)
  }

  config.Sources[name] = ian.CalendarSource{
  	Source:   source,
  	Type:     _type,
  	Lifetime: lifetime,
  }

  if err := ian.WriteConfig(GetRoot(), config); err != nil {
    log.Fatal(err)
  }

  fmt.Printf("source '%s' added\n", name)
}
