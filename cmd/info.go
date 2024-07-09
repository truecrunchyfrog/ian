package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

var index int

func init() {
  infoCmd.Flags().IntVarP(&index, "index", "i", 0, "Choose an event based on index for ambiguous results.")

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
	case len(matches) > 1 && index == 0:
		fmt.Printf("ambiguous query gave %d results:\n\n", len(matches))
		for i, match := range matches {
			fmt.Printf("[%d] %s\n", i+1, match.Path)
		}
    fmt.Println("\nappend '-i n' to the command to show info for result n.")
	default:
    if index < 0 || index > len(matches) {
      log.Fatalf("cannot use index '%d' on search with '%d' result(s)\n", index, len(matches))
    }
    if index > 0 {
      index--
    }
		event := matches[index] // index defaults to 0 so it will be valid even for unambiguous results

    pairs := []struct{key string; value any}{
      { "path", event.Path },
      { "cached", event.Cached },
      { "", "" },
      { "summary", event.Props.Summary },
      { "start", event.Props.Start },
      { "end", event.Props.End },
      { "duration", ian.DurationToString(event.Props.End.Sub(event.Props.Start)) },
      { "all-day", event.Props.AllDay },
      { "", "" },
      { "description", event.Props.Description },
      { "location", event.Props.Location },
      { "url", event.Props.Url },
      { "", "" },
      { "created", event.Props.Created },
      { "modified", event.Props.Modified },
    }

    for _, keyValue := range pairs {
      fmt.Println(DisplayKeyValue(keyValue.key, keyValue.value))
    }
	}
}

func DisplayKeyValue(key string, value any) string {
  return fmt.Sprintf("\033[2m%-10s\033[0m %v", key, value)
}
