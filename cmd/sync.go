package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

func init() {
	rootCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Dispatch a synchronization event.",
	Long:  "Ping the synchronization listeners. Useful for unnoticed event changes, like manual changes or during a cooldown.",
	Args:  cobra.NoArgs,
	Run:   syncCmdRun,
}

func syncCmdRun(cmd *cobra.Command, args []string) {
	instance, err := ian.CreateInstance(GetRoot(), false)
	if err != nil {
		log.Fatal(err)
	}

  fmt.Println("syncing...")

  if err := instance.Sync(ian.SyncEvent{
		Type:    ian.SyncEventPing,
    Message: "ian: manual sync",
  }); err != nil {
    log.Fatal(err)
  }
}
