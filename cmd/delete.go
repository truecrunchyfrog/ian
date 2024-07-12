package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

func init() {
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
    log.Fatal("no such event")
  }
  if len(matches) > 1 {
    log.Fatal("ambiguous event")
  }

  event := matches[0]

  if event.Constant {
    log.Fatalf("'%s' is a constant event and cannot be deleted by itself.\n", event.Path)
  }

  if err := os.Remove(event.GetRealPath(instance)); err != nil {
    log.Fatal(err)
  }

  fmt.Println("deleted event")

  instance.Sync(ian.SyncEvent{
    Type: ian.SyncEventDelete,
    Files: event.GetRealPath(instance),
    Message: fmt.Sprintf("ian: delete event '%s'", event.Path),
  }, false, nil)
}
