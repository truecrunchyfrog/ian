package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

var deleteMatchIndex int
var deleteAllMatches bool

func init() {
	deleteCmd.Flags().IntVarP(&deleteMatchIndex, "index", "i", 0, "Delete match by `index`.")
	deleteCmd.Flags().BoolVarP(&deleteAllMatches, "all", "A", false, "Delete all matches.")
	deleteCmd.MarkFlagsMutuallyExclusive("index", "all")

	eventPropsCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "rm <event>",
	Short: "Delete an event",
	Args:  cobra.ExactArgs(1),
	Run:   deleteCmdRun,
}

func deleteCmdRun(cmd *cobra.Command, args []string) {
	instance, err := ian.CreateInstance(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

	events, err := instance.ReadEvents(ian.TimeRange{})
	if err != nil {
		log.Fatal(err)
	}

	matches := ian.QueryEvents(&events, args[0])

	if len(matches) == 0 {
		log.Fatal("no event matched query")
	}

	fmt.Printf("matched %d event(s):\n", len(matches))
	for i, match := range matches {
		fmt.Printf("%3s[%d] %s\n", "", i+1, match.Path)
	}
	fmt.Println()

	if len(matches) > 1 && deleteMatchIndex == 0 && !deleteAllMatches {
		log.Fatal("ambiguous event. append the flag '-i N' to delete match N, or '--all' to delete all matches.")
	}

	var deleteEvents []ian.Event
	switch {
	case deleteMatchIndex < 0 || deleteMatchIndex > len(matches):
		log.Fatalf("index %d out of bounds for %d matches\n", deleteMatchIndex, len(matches))
	case deleteMatchIndex != 0:
		deleteEvents = []ian.Event{matches[deleteMatchIndex-1]}
	case deleteAllMatches:
		deleteEvents = matches
  default:
    deleteEvents = []ian.Event{matches[0]}
	}

	for _, deleteEvent := range deleteEvents {
		if deleteEvent.Constant {
			log.Fatalf("'%s' is a constant event and cannot be deleted by itself.\n", deleteEvent.Path)
		}
	}

	syncMsg := "ian: delete "
	if len(deleteEvents) > 1 {
		syncMsg += fmt.Sprintf("%d events", len(deleteEvents))
	} else {
		syncMsg += "event"
	}

	deletedPaths := []string{}
	deletedFiles := []string{}

	for _, deleteEvent := range deleteEvents {
		file := deleteEvent.GetRealPath(instance)
		deletedFiles = append(deletedFiles, file)
		if err := os.Remove(file); err != nil {
			log.Fatal(err)
		}

		deletedPaths = append(deletedPaths, "'"+deleteEvent.Path+"'")

		fmt.Printf("deleted event '%s'\n", deleteEvent.Path)
	}

	syncMsg += " " + strings.Join(deletedPaths, ", ")

	instance.Sync(ian.SyncEvent{
		Type:    ian.SyncEventDelete,
		Files:   strings.Join(deletedFiles, " "),
		Message: syncMsg,
	}, false, nil)
}
