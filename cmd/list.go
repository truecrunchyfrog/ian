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
	listCmd.Flags().BoolVarP(&showNow, "now", "n", false, "Show an event to indicate today.")
	listCmd.Flags().BoolVarP(&showPast, "past", "p", false, "Show past events.")

	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List events",
	Args:  cobra.NoArgs,
	Run:   listCmdRun,
}

func listCmdRun(cmd *cobra.Command, args []string) {
  instance, err := ian.CreateInstance(GetRoot(), true)
  if err != nil {
    log.Fatal(err)
  }

	now := time.Now()
  events := slices.Clone(instance.Events)

	if !showPast {
		events = slices.DeleteFunc(events, func(event ian.Event) bool {
			return event.Props.End.Before(now)
		})
	}

	if showNow {
		events = append(events, ian.Event{
			Props: ian.EventProperties{
				Summary: "--- now ---",
				Start:   now,
				End:     now,
			},
			Constant: true,
		})
	}

	slices.SortFunc(events, func(a ian.Event, b ian.Event) int {
		switch {
		case a.Props.Start.Before(b.Props.Start):
			return -1
		case a.Props.Start.After(b.Props.Start):
			return 1
		default:
			return 0
		}
	})

	for _, event := range events {
		fmt.Println(event.String())
	}
}
