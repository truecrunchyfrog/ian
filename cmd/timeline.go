package cmd

import (
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

func init() {
	timelineCmd.Flags().BoolP("past", "p", false, "Show past events")
	timelineCmd.Flags().StringSliceP("calendars", "c", nil, "Limit the shown events to those contained in the calendars in this `list`.")
	timelineCmd.Flags().Bool("no-legend", false, "Do not show the calendar legend that shows what colors belong to what calendar.")

	rootCmd.AddCommand(timelineCmd)
}

var timelineCmd = &cobra.Command{
	Use:     "timeline [from [to]]",
	Aliases: []string{"time", "t", "tl", "events", "evs"},
	Short:   "View events in a timeline",
	Long:    "View events in a beautiful linear timeline. Without any arguments, 'from' is now, and 'to' is 5 years ahead in time. Works good with 'more' and 'less'.",
	Args:    cobra.RangeArgs(0, 2),
	Run:     timelineCmdRun,
}

func timelineCmdRun(cmd *cobra.Command, args []string) {
	instance, err := ian.CreateInstance(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

	var timeRange ian.TimeRange

	if len(args) >= 1 {
		timeRange.From, err = ian.ParseDateTime(args[0], ian.GetTimeZone())
		if err != nil {
			log.Fatal(err)
		}
	} else {
		now := time.Now().In(ian.GetTimeZone())
		timeRange.From = time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			0, 0, 0, 0,
			ian.GetTimeZone(),
		)
		timeRange.To = timeRange.From.AddDate(5, 0, 0)
	}

	if len(args) >= 2 {
		timeRange.To, err = ian.ParseDateTime(args[1], ian.GetTimeZone())
		if err != nil {
			log.Fatal(err)
		}
	}

	events, unsatisfiedRecurrences, _ := instance.ReadEvents(timeRange)

	if cals, _ := cmd.Flags().GetStringSlice("calendars"); len(cals) != 0 {
		events = ian.FilterEvents(&events, func(e *ian.Event) bool {
			return slices.Contains(cals, e.Path.Calendar())
		})
	}

	if len(events) == 0 {
		log.Fatal("no events to show!")
	}

	fmt.Println(ian.DisplayTimeline(instance, events, ian.GetTimeZone()))
	fmt.Println(ian.DisplayUnsatisfiedRecurrences(instance, unsatisfiedRecurrences))

	if hide, _ := cmd.Flags().GetBool("no-legend"); !hide {
		fmt.Println("\n" + ian.DisplayCalendarLegend(instance, events))
	}
}
