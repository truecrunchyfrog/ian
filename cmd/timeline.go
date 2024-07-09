package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

func init() {
	rootCmd.AddCommand(timelineCmd)
}

var timelineCmd = &cobra.Command{
	Use:   "timeline <from> <to>",
	Short: "View events in a timeline",
	Args:  cobra.ExactArgs(2),
	Run:   timelineCmdRun,
}

func timelineCmdRun(cmd *cobra.Command, args []string) {
  instance, err := ian.CreateInstance(GetRoot(), true)
  if err != nil {
    log.Fatal(err)
  }

  from, err := ian.ParseDateTime(args[0], GetTimeZone())
  if err != nil {
    log.Fatal(err)
  }
  to, err := ian.ParseDateTime(args[1], GetTimeZone())
  if err != nil {
    log.Fatal(err)
  }

  fmt.Println(ian.DisplayTimeline(instance, from, to, instance.Events, GetTimeZone()))
}
