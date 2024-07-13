package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

var index int

func init() {
  infoCmd.Flags().IntVarP(&index, "index", "i", 0, "Choose an event based on index for ambiguous results.")

	eventPropsCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info <event>",
	Short: "View event info",
	Args:  cobra.ExactArgs(1),
	Run:   infoCmdRun,
}

func infoCmdRun(cmd *cobra.Command, args []string) {
	instance, err := ian.CreateInstance(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

  events, _, err := instance.ReadEvents(ian.TimeRange{})

	query := args[0]
  matches := ian.QueryEvents(&events, query)

	switch {
	case len(matches) < 1:
		fmt.Printf("no event matches query: '%s'\n", query)
	case len(matches) > 1 && index == 0:
		fmt.Printf("ambiguous query gave %d results:\n\n", len(matches))
		for i, match := range matches {
			fmt.Printf("[%d] %s\n", i+1, match.Path)
		}
    fmt.Println("\nappend '-i N' to the command to show info for result n.")
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
      { "constant", event.Constant },
      { "parent", event.Parent },
      { "", "" },
      { "summary", event.Props.Summary },
      { "start", event.Props.Start },
      { "end", event.Props.End },
      { "duration", ian.DurationToString(event.Props.End.Sub(event.Props.Start)) },
      { "all-day", event.Props.AllDay },
      { "recurrence", event.Props.Rrule },
      { "", "" },
      { "description", event.Props.Description },
      { "location", event.Props.Location },
      { "url", event.Props.Url },
      { "", "" },
      { "created", event.Props.Created },
      { "modified", event.Props.Modified },
      { "", "" },
      { "uid", event.Props.Uid },
    }

    for _, keyValue := range pairs {
      fmt.Println(DisplayKeyValue(keyValue.key, keyValue.value))
    }
	}
}

func DisplayKeyValue(key string, value any) string {
  return fmt.Sprintf("\033[2m%-10s\033[0m %v", key, value)
}
