package main

import (
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

var showNow bool
var showPast bool

func init() {
	listCommand.Flags().BoolVarP(&showNow, "now", "n", false, "Show an event to indicate today.")
	listCommand.Flags().BoolVarP(&showPast, "past", "p", false, "Show past events.")

	rootCmd.AddCommand(listCommand)
}

var listCommand = &cobra.Command{
	Use:   "list",
	Short: "List events",
	Args:  cobra.NoArgs,
	Run:   listCommandRun,
}

func listCommandRun(cmd *cobra.Command, args []string) {
	allEvents, err := ian.ReadEventDir(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

  now := time.Now()

  if !showPast {
    allEvents = slices.DeleteFunc(allEvents, func(event ian.CalendarEvent) bool {
      return event.End.Before(now)
    })
  }

	if showNow {
		allEvents = append(allEvents, ian.CalendarEvent{
			Summary:  "--- now ---",
			Start: now,
			End:   now,
		})
	}

	slices.SortFunc(allEvents, func(a ian.CalendarEvent, b ian.CalendarEvent) int {
		switch {
		case a.Start.Before(b.Start):
			return -1
		case a.Start.After(b.Start):
			return 1
		default:
			return 0
		}
	})

	for _, event := range allEvents {
		fmt.Println(event.String())
	}
}
