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
	allEvents, err := ian.ReadEvents(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now()

	if !showPast {
		allEvents = slices.DeleteFunc(allEvents, func(event ian.Event) bool {
			return event.Properties.End.Before(now)
		})
	}

	if showNow {
		allEvents = append(allEvents, ian.Event{
			Properties: ian.EventProperties{
				Summary: "--- now ---",
				Start:   now,
				End:     now,
			},
			Constant: true,
		})
	}

	slices.SortFunc(allEvents, func(a ian.Event, b ian.Event) int {
		switch {
		case a.Properties.Start.Before(b.Properties.Start):
			return -1
		case a.Properties.Start.After(b.Properties.Start):
			return 1
		default:
			return 0
		}
	})

	for _, event := range allEvents {
		fmt.Println(event.String())
	}
}
