package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

func init() {
	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info <query>",
	Short: "View event info",
	Args:  cobra.ExactArgs(1),
	Run:   infoCmdRun,
}

func infoCmdRun(cmd *cobra.Command, args []string) {
	instance, err := ian.CreateInstance(GetRoot(), true)
	if err != nil {
		log.Fatal(err)
	}

	query := args[0]
	var matches []ian.Event

	for _, ev := range instance.Events {
		if strings.Contains(ev.Path, query) || strings.Contains(ev.Props.Summary, query) || strings.Contains(ev.Props.Description, query) {
			matches = append(matches, ev)
		}
	}

	switch {
	case len(matches) < 1:
		fmt.Printf("no loaded event matches query: '%s'\n", query)
	case len(matches) > 1:
		fmt.Printf("ambiguous query gave %d results:\n", len(matches))
		for _, match := range matches {
			fmt.Println(match.Path)
		}
	default:
		event := matches[0]

    fmt.Printf("%s\n\nfile: %s\n\nsummary: %s\nproperties: %#v\n", event.Path, event.GetRealPath(instance), event.Props.Summary, event.Props)
	}
}
