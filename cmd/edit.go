package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

const renameFlag string = "rename"
const copyFlag string = "copy"
const updateUidFlag string = "update-uid"

var editFlags []string = []string{
  renameFlag,
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

func init() {
	editCmd.Flags().String(renameFlag, "", "Rename the event.")
	editCmd.Flags().String(copyFlag, "", "Copy the event to the `destination` path, along with the changes.")
	editCmd.Flags().Bool(updateUidFlag, false, "Update the UID of an event.")

	eventPropsCmd.AddCommand(editCmd)
}

var editCmd = &cobra.Command{
	Use:     "edit event...",
	Aliases: []string{"e", "ed", "ch", "m", "mod", "modify"},
	Short:   "Edit an event's properties",
	Run:     editCmdRun,
}

func editCmdRun(cmd *cobra.Command, args []string) {
	var modified []string
	for _, flag := range editFlags {
		if cmd.Flags().Changed(flag) {
			modified = append(modified, flag)
		}
	}

	if len(modified) == 0 {
		log.Fatal("no modifications. check the help page for a list of values to change.")
	}

	instance, err := ian.CreateInstance(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

	events, _, err := instance.ReadEvents(ian.TimeRange{})
	if err != nil {
		log.Fatal(err)
	}

	editEvents := []*ian.Event{}

	if len(args) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			args = append(args, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}

	for _, arg := range args {
		event, err := ian.GetEvent(&events, arg)
		if err != nil {
			log.Fatal(err)
		}
		editEvents = append(editEvents, event)
	}

	if len(args) == 0 {
		log.Fatal("no event to edit")
	}

	onWritten := []func(){}

	files := []string{}

  syncMsg := "ian: edit "
  if len(editEvents) > 1 {
    syncMsg += fmt.Sprintf("%d events; ", len(editEvents))
  } else {
    syncMsg += "event: "
  }

	for i, event := range editEvents {
		if event.Constant {
			log.Fatalf("'%s' is a constant event and cannot be modified.\n", event.Path)
		}

		files = append(files, event.GetFilepath(instance))
    if i != 0 {
      syncMsg += ", "
    }
    syncMsg += "'" + event.Path + "'"

    if cmd.Flags().Changed(renameFlag) { // Rename operation
      if len(editEvents) > 1 {
        log.Fatal("rename cannot be used on multiple events")
      }
      newName, _ := cmd.Flags().GetString(renameFlag)
      newPath := ian.SanitizePath(path.Join(path.Dir(event.Path), newName))
      if newPath == event.Path {
        log.Fatal("rename is trying to rename to the current name")
      }
      if _, err := ian.GetEvent(&events, newPath); err == nil {
				log.Fatalf("a file with the path '%s' already exists.\n", newPath)
      }
			if ian.IsPathInCache(newPath) {
				log.Fatal("cannot set calendar to inside cache")
			}
			onWritten = append(onWritten, func() {
				os.Remove(event.Path)
			})
			log.Printf("note: '%s' is being moved to '%s'.\n", event.Path, newPath)

      event.Path = newPath
    }
		if cmd.Flags().Changed(copyFlag) { // Copy operation
      if len(editEvents) > 1 {
        log.Fatal("copy cannot be used on multiple events")
      }
			event.Path, _ = cmd.Flags().GetString(copyFlag)
			if _, err := ian.GetEvent(&events, event.Path); err == nil {
				log.Fatalf("a file with the path '%s' already exists.\n", event.Path)
			}
			if ian.IsPathInCache(event.Path) {
				log.Fatal("cannot set calendar to inside cache")
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
      oldPath := event.Path
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
			log.Printf("note: '%s' is being moved to '%s'.\n", oldPath, event.Path)
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

		event.Props.Modified = time.Now().In(ian.GetTimeZone())

		if err := event.Props.Validate(); err != nil {
			log.Fatalf("validation failed: %s", err)
		}

		checkCollision(&events, event.Props)
	}

  syncMsg += "; " + strings.Join(modified, ", ")

	err = instance.Sync(func() error {
		for _, event := range editEvents {
			if err := event.Write(instance); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("'%s' has been updated; %s\n", event.Path, strings.Join(modified, ", "))
		}
		for _, f := range onWritten {
			f()
		}
		return nil
	}, ian.SyncEvent{
		Type:    ian.SyncEventUpdate,
		Files:   files,
		Message: syncMsg,
	}, false, nil)

	if err != nil {
		log.Fatal(err)
	}
}
