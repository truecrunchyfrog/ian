package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/emersion/go-ical"
	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

func init() {
	migrateImportCmd.Flags().StringP("dir", "d", "", "Directory to write the event files to.")

	migrateCmd.AddCommand(migrateImportCmd)
}

var migrateImportCmd = &cobra.Command{
	Use:     "import -d dir",
	Short:   "Import from iCalendar.",
	Example:
  `# extract all events from 'cal.ics' as event files and put them in the directory 'cal':
cat cal.ics | ian migrate import -d cal`,
	Args:    cobra.NoArgs,
	Run:     migrateImportCmdRun,
}

func migrateImportCmdRun(cmd *cobra.Command, args []string) {
	ics, err := ical.NewDecoder(os.Stdin).Decode()
	if err != nil {
		log.Fatal(err)
	}

	propsList, err := ian.FromIcal(ics)
	if err != nil {
		log.Fatal(err)
	}

	dest, _ := cmd.Flags().GetString("dir")
  if err := ian.CreateDir(dest); err != nil {
    log.Fatal(err)
  }
	for _, props := range propsList {
		props.Write(filepath.Join(dest, props.FormatName()))
	}
}
