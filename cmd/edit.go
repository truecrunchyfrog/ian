package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

const copyFlag string = "copy"
const updateUidFlag string = "update-uid"

func init() {
  editCmd.Flags().String(copyFlag, "", "Copy the event to the `destination` path, along with the changes.")
  editCmd.Flags().Bool(updateUidFlag, false, "Update the UID of an event.")

	eventPropsCmd.AddCommand(editCmd)
}

var editCmd = &cobra.Command{
	Use:     "edit <event>",
	Aliases: []string{"ed", "ch"},
	Short:   "Edit an event's properties",
	Args:    cobra.ExactArgs(1),
	Run:     editCmdRun,
}

func editCmdRun(cmd *cobra.Command, args []string) {
	instance, err := ian.CreateInstance(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

	events, _, err := instance.ReadEvents(ian.TimeRange{})
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
		log.Fatalf("'%s' is a constant event and cannot be modified.\n", event.Path)
	}

	onWritten := []func(){}

	flags := []string{
    copyFlag,
    updateUidFlag,
		eventFlag_Summary,
		eventFlag_Start,
		eventFlag_End,
		eventFlag_Description,
		eventFlag_Location,
		eventFlag_Url,
		eventFlag_Duration,
		eventFlag_Hours,
		eventFlag_Calendar,
		eventFlag_Rrule,
		eventFlag_Rdate,
		eventFlag_ExDate,
	}

  if cmd.Flags().Changed(copyFlag) { // Copy operation
    event.Path, _ = cmd.Flags().GetString(copyFlag)
    if e, _ := ian.GetEvent(&events, event.Path); e != nil {
      log.Fatalf("a file with the path '%s' already exists.\n", event.Path)
    }
  }
  if cmd.Flags().Changed(updateUidFlag) { // UID
    event.Props.Uid = ian.GenerateUid()
  }
	if eventFlags.Changed(eventFlag_Summary) { // Summary
		event.Props.Summary, _ = eventFlags.GetString(eventFlag_Summary)
	}
	if eventFlags.Changed(eventFlag_Start) { // Start
		startString, _ := eventFlags.GetString(eventFlag_Start)
		start, err := ian.ParseDateTime(startString, ian.GetTimeZone())
		if err != nil {
			log.Fatal(err)
		}
		event.Props.Start = start
	}
	if eventFlags.Changed(eventFlag_End) { // End
		endString, _ := eventFlags.GetString(eventFlag_End)
		end, err := ian.ParseDateTime(endString, ian.GetTimeZone())
		if err != nil {
			log.Fatal(err)
		}
		event.Props.End = end
	}
	if eventFlags.Changed(eventFlag_Description) { // Description
		event.Props.Description, _ = eventFlags.GetString(eventFlag_Description)
	}
	if eventFlags.Changed(eventFlag_Location) { // Location
		event.Props.Location, _ = eventFlags.GetString(eventFlag_Location)
	}
	if eventFlags.Changed(eventFlag_Url) { // URL
		event.Props.Url, _ = eventFlags.GetString(eventFlag_Url)
	}
	if eventFlags.Changed(eventFlag_Duration) { // Duration
		d, _ := eventFlags.GetDuration(eventFlag_Duration)
		event.Props.End = event.Props.Start.Add(d)
	}
	if eventFlags.Changed(eventFlag_Hours) { // Hours
		hours, _ := eventFlags.GetStringSlice(eventFlag_Hours)
		if err := handleHours(hours, &event.Props.Start, &event.Props.End); err != nil {
			log.Fatal(err)
		}
	}
	if eventFlags.Changed(eventFlag_Calendar) { // Calendar (move operation)
		oldFile := event.GetFilepath(instance)
		newCalendar, _ := eventFlags.GetString(eventFlag_Calendar)
		event.Path = path.Join(newCalendar, path.Base(event.Path))
    if e, _ := ian.GetEvent(&events, event.Path); e != nil {
      log.Fatalf("a file with the path '%s' already exists.\n", event.Path)
    }
		if ian.IsPathInCache(event.Path) {
			log.Fatal("cannot set calendar to inside cache")
		}
		onWritten = append(onWritten, func() {
			os.Remove(oldFile)
		})
		fmt.Println("note: event is being moved to another file location.")
	}
	if eventFlags.Changed(eventFlag_Rrule) { // Rrule
		recurrenceFlag, _ := eventFlags.GetString(eventFlag_Rrule)
		event.Props.Recurrence.RRule = recurrenceFlag
	}
	if eventFlags.Changed(eventFlag_Rdate) { // Rdate
		recurrenceFlag, _ := eventFlags.GetString(eventFlag_Rdate)
		event.Props.Recurrence.RDate = recurrenceFlag
	}
	if eventFlags.Changed(eventFlag_ExDate) { // ExDate
		recurrenceFlag, _ := eventFlags.GetString(eventFlag_ExDate)
		event.Props.Recurrence.ExDate = recurrenceFlag
	}

	var modified []string
	for _, flag := range flags {
		if cmd.Flags().Changed(flag) {
			modified = append(modified, flag)
		}
	}

	if len(modified) == 0 {
		log.Fatal("no changes described. check the help page for a list of values to change.")
	}
	event.Props.Modified = time.Now().In(ian.GetTimeZone())
	if err := event.Props.Validate(); err != nil {
		log.Fatalf("verification failed: %s", err)
	}

	events2, _, err := instance.ReadEvents(event.Props.GetTimeRange())
	if err != nil {
		log.Fatal(err)
	}

	checkCollision(&events2, event.Props)

	if err := event.Write(instance); err != nil {
		log.Fatal(err)
	}
	for _, f := range onWritten {
		f()
	}
	fmt.Printf("'%s' has been updated; %s\n", event.Path, strings.Join(modified, ", "))

	instance.Sync(ian.SyncEvent{
		Type:    ian.SyncEventUpdate,
		Files:   event.GetFilepath(instance),
		Message: fmt.Sprintf("ian: edit event '%s'; %s", event.Path, strings.Join(modified, ", ")),
	}, false, nil)
}
