package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

func init() {
	rootCmd.AddCommand(findCmd)
}

var findCmd = &cobra.Command{
	Use:   "find query",
	Short: "Query matches",
	Run:   findCmdRun,
}

func findCmdRun(cmd *cobra.Command, args []string) {
	instance, err := ian.CreateInstance(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

	events, _, err := instance.ReadEvents(ian.TimeRange{})
	if err != nil {
		log.Fatal(err)
	}

	query = strings.ToLower(query)
	return FilterEvents(events, func(e *Event) bool {
		return strings.Contains(strings.ToLower(e.Path), query) ||
			strings.Contains(strings.ToLower(e.Props.Summary), query) ||
			strings.Contains(strings.ToLower(e.Props.Description), query)
	})
}
