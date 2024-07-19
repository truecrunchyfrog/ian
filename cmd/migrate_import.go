package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	migrateCmd.AddCommand(migrateImportCmd)
}

var migrateImportCmd = &cobra.Command{
	Use:   "import file...",
	Short: "Import from iCalendar.",
	Run:   migrateImportCmdRun,
}

func migrateImportCmdRun(cmd *cobra.Command, args []string) {
  panic("TODO: implement")
}
