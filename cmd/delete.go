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
	Use:   "rm <path>",
	Short: "Delete an event",
	Args:  cobra.ExactArgs(1),
	Run:   deleteCmdRun,
}

func deleteCmdRun(cmd *cobra.Command, args []string) {
  instance, err := ian.CreateInstance(GetRoot(), true)
  if err != nil {
    log.Fatal(err)
  }

  event, err := instance.GetEvent(args[0])
  if err != nil {
    log.Fatal(err)
  }

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
