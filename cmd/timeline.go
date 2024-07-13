package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

var noLegend bool

func init() {
	timelineCmd.Flags().BoolVar(&noLegend, "no-legend", false, "Do not show the calendar legend that shows what colors belong to what calendar.")

	rootCmd.AddCommand(timelineCmd)
}

var timelineCmd = &cobra.Command{
	Use:   "timeline [from [to]]",
	Short: "View events in a timeline",
  Long: "View all events in a visual timeline. 'from' defaults to year 0. 'to' defaults to 5 years ahead in time.",
	Args:  cobra.RangeArgs(0, 2),
	Run:   timelineCmdRun,
}

func timelineCmdRun(cmd *cobra.Command, args []string) {
	instance, err := ian.CreateInstance(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

  var timeRange ian.TimeRange

	if len(args) >= 1 {
		timeRange.From, err = ian.ParseDateTime(args[0], GetTimeZone())
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(args) >= 2 {
		timeRange.To, err = ian.ParseDateTime(args[1], GetTimeZone())
		if err != nil {
			log.Fatal(err)
		}
	}

  events, unsatisfiedRecurrences, _ := instance.ReadEvents(timeRange)

	fmt.Println(ian.DisplayTimeline(instance, events, GetTimeZone()))
  fmt.Println(ian.DisplayUnsatisfiedRecurrences(instance, unsatisfiedRecurrences))

	if !noLegend {
		fmt.Println("\n" + ian.DisplayCalendarLegend(instance, events))
	}
}
