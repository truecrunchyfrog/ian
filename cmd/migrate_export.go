package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

var includeCache bool

var cherrypickCalendars []string
var cherrypickEvents []string

var excludeCalendars []string
var excludeEvents []string

func init() {
	migrateExportCmd.Flags().StringP("file", "f", "", "Export the events to a file.")
	migrateExportCmd.Flags().StringP("directory", "d", "", "Export the events, with each calendars in its own file.")

	migrateExportCmd.MarkFlagsMutuallyExclusive("file", "directory")

	migrateExportCmd.Flags().BoolVar(&includeCache, "include-cache", false, "Include cached events from sources in the export.")
	migrateExportCmd.Flags().StringSliceVar(&cherrypickCalendars, "cherrypick-calendars", nil, "Include nothing but the events in these calendars.")
	migrateExportCmd.Flags().StringSliceVar(&cherrypickEvents, "cherrypick-events", nil, "Include nothing but these events.")
	migrateExportCmd.Flags().StringSliceVar(&excludeCalendars, "exclude-calendars", nil, "Include everything except the events in these calendars.")
	migrateExportCmd.Flags().StringSliceVar(&excludeEvents, "exclude-events", nil, "Include everything except these events.")

	migrateExportCmd.MarkFlagsMutuallyExclusive(
		"include-cache",
		"cherrypick-calendars",
		"cherrypick-events",
		"exclude-calendars",
		"exclude-events",
	)

	migrateCmd.AddCommand(migrateExportCmd)
}

var migrateExportCmd = &cobra.Command{
	Use:   "export [-f file | -d dir]",
	Short: "Export to iCalendar.",
	Long:  "If both file and directory are left out, all output is sent to stdout.",
	Run:   migrateExportCmdRun,
}

func migrateExportCmdRun(cmd *cobra.Command, args []string) {
	var filterFunc func(e *ian.Event) bool

	switch {
	case includeCache:
		filterFunc = func(e *ian.Event) bool {
			return true
		}
	case cherrypickCalendars != nil:
		filterFunc = func(e *ian.Event) bool {
			// Only from these calendars.
			return slices.Contains(cherrypickCalendars, e.GetCalendarName())
		}
	case cherrypickEvents != nil:
		filterFunc = func(e *ian.Event) bool {
			// Only these events.
			return slices.Contains(cherrypickEvents, e.Path)
		}
	case excludeCalendars != nil:
		filterFunc = func(e *ian.Event) bool {
			// NOT these calendars.
			return !slices.Contains(excludeCalendars, e.GetCalendarName())
		}
	case excludeEvents != nil:
		filterFunc = func(e *ian.Event) bool {
			// NOT these events.
			return !slices.Contains(cherrypickEvents, e.Path)
		}
	default:
		filterFunc = func(e *ian.Event) bool {
			return e.Type != ian.EventTypeCache
		}
	}

	instance, err := ian.CreateInstance(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

	events, _, err := instance.ReadEvents(ian.TimeRange{})
	if err != nil {
		log.Fatal(err)
	}

	events = ian.FilterEvents(&events, func(e *ian.Event) bool {
		return e.Type != ian.EventTypeRecurrence && filterFunc(e)
	})

	if cmd.Flags().Changed("file") {
		fileDest, _ := cmd.Flags().GetString("file")
		ics := ian.ToIcal(events)
		if err := os.WriteFile(fileDest, []byte(ics.Serialize()), 0644); err != nil {
			log.Fatal(err)
		}
	} else if cmd.Flags().Changed("directory") {
		dirDest, _ := cmd.Flags().GetString("directory")
		ian.CreateDir(dirDest)

		eventsByCal := map[string][]ian.Event{}

		for _, event := range events {
			calEvents := eventsByCal[event.GetCalendarName()]
			if calEvents == nil {
				calEvents = []ian.Event{}
			}
			calEvents = append(calEvents, event)
		}

		for cal, events := range eventsByCal {
			filename := strings.NewReplacer(
				"-", "",
				"/", "-",
			).Replace(cal) + ".ics"
			ics := ian.ToIcal(events)
			if err := os.WriteFile(filepath.Join(dirDest, filename), []byte(ics.Serialize()), 0644); err != nil {
				log.Fatal(err)
			}
		}
	} else {
		fmt.Print(ian.ToIcal(events).Serialize())
	}
}