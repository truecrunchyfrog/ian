package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

var ignoreCooldowns bool
var listListeners bool

func init() {
  syncCmd.Flags().BoolVarP(&ignoreCooldowns, "ignore-cooldowns", "i", false, "Ignore any listener cooldowns.")
  syncCmd.Flags().BoolVarP(&listListeners, "list", "l", false, "List configured sync listeners instead of syncing.")

	rootCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Dispatch a synchronization event.",
	Long:  "Ping the synchronization listeners. Useful for unnoticed event changes, like manual changes or during a cooldown with '--ignore-cooldowns'. Also prints stdout.",
	Args:  cobra.NoArgs,
	Run:   syncCmdRun,
}

func syncCmdRun(cmd *cobra.Command, args []string) {
	instance, err := ian.CreateInstance(GetRoot(), false)
	if err != nil {
		log.Fatal(err)
	}

  if listListeners {
    fmt.Println("configured sync listeners:")
    for name, listener := range instance.Config.Sync.Listeners {
      fmt.Printf("'%s' has command '%s' with a cooldown of %s\n", name, listener.Command, listener.Cooldown_)
    }
    fmt.Println("\nsync is not made when listing listeners.")
    return
  }

  fmt.Println("syncing...\n")

  if err := instance.Sync(ian.SyncEvent{
		Type:    ian.SyncEventPing,
    Message: "ian: manual sync",
  }, ignoreCooldowns, os.Stdout); err != nil {
    log.Fatal(err)
  }

  fmt.Println("sync done")
}
