package cmd

import (
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

var noLegend bool

func init() {
	timelineCmd.Flags().BoolVar(&noLegend, "no-legend", false, "Do not show the calendar legend that shows what colors belong to what calendar.")

	rootCmd.AddCommand(timelineCmd)
}

var timelineCmd = &cobra.Command{
	Use:   "timeline [<from> [to]]",
	Short: "View events in a timeline",
	Args:  cobra.RangeArgs(0, 2),
	Run:   timelineCmdRun,
}

func timelineCmdRun(cmd *cobra.Command, args []string) {
	instance, err := ian.CreateInstance(GetRoot(), true)
	if err != nil {
		log.Fatal(err)
	}

	events := slices.Clone(instance.Events)

	if len(args) >= 1 {
		var from, to time.Time

		from, err = ian.ParseDateTime(args[0], GetTimeZone())
		if err != nil {
			log.Fatal(err)
		}
		if len(args) >= 2 {
			to, err = ian.ParseDateTime(args[1], GetTimeZone())
			if err != nil {
				log.Fatal(err)
			}
		}

    events = slices.DeleteFunc(events, func(event *ian.Event) bool {
      if to.IsZero() {
        return event.Props.Start.Before(from)
      }
      return !ian.IsPeriodConfinedToPeriod(event.Props.Start, event.Props.End, from, to)
    })
	}

	fmt.Println(ian.DisplayTimeline(instance, events, GetTimeZone()))

	if !noLegend {
		fmt.Println("\n" + ian.DisplayCalendarLegend(instance, events))
	}
}
