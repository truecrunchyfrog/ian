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
  instance, err := ian.CreateInstance(GetRoot(), true)
  if err != nil {
    log.Fatal(err)
  }

	now := time.Now()
  events := slices.Clone(instance.Events)

	if !showPast {
		events = slices.DeleteFunc(events, func(event ian.Event) bool {
			return event.Properties.End.Before(now)
		})
	}

	if showNow {
		events = append(events, ian.Event{
			Properties: ian.EventProperties{
				Summary: "--- now ---",
				Start:   now,
				End:     now,
			},
			Constant: true,
		})
	}

	slices.SortFunc(events, func(a ian.Event, b ian.Event) int {
		switch {
		case a.Properties.Start.Before(b.Properties.Start):
			return -1
		case a.Properties.Start.After(b.Properties.Start):
			return 1
		default:
			return 0
		}
	})

	for _, event := range events {
		fmt.Println(event.String())
	}
}
