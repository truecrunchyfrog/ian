package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

func init() {
	rootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete <path>",
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

  if event.Cached {
    log.Fatalf("'%s' is a cached event. to remove the entire source, remove it from your configuration and run 'ian sources --clean'. alternatively delete the file manually (not recommended).\n", event.Path)
  }

  if err := os.Remove(event.GetRealPath(instance)); err != nil {
    log.Fatal(err)
  }

  fmt.Println("deleted event")
}
