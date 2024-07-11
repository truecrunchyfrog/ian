package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/teambition/rrule-go"
	"github.com/truecrunchyfrog/ian"
)

func init() {
	eventPropsCmd.AddCommand(editCmd)
}

var editCmd = &cobra.Command{
	Use:   "edit <path>",
	Short: "Edit an event's properties",
	Args:  cobra.ExactArgs(1),
	Run:   editCmdRun,
}

func editCmdRun(cmd *cobra.Command, args []string) {
	instance, err := ian.CreateInstance(GetRoot(), true)
	if err != nil {
		log.Fatal(err)
	}

	event, err := instance.GetEvent(args[0])
	if err != nil {
		log.Fatal(err)
	}

	if event.Constant {
		log.Fatalf("'%s' is a constant event and cannot be modified.\n", event.Path)
	}

  flags := []string {
    eventFlag_Summary,
    eventFlag_Start,
    eventFlag_End,
    eventFlag_AllDay,
    eventFlag_Description,
    eventFlag_Location,
    eventFlag_Url,
    eventFlag_Duration,
    eventFlag_Hours,
    eventFlag_Calendar,
    eventFlag_Recurrence,
  }

	if eventFlags.Changed(eventFlag_Summary) { // Summary
		event.Props.Summary, _ = eventFlags.GetString(eventFlag_Summary)
  }
	if eventFlags.Changed(eventFlag_Start) { // Start
		startString, _ := eventFlags.GetString(eventFlag_Start)
		start, err := ian.ParseDateTime(startString, GetTimeZone())
		if err != nil {
			log.Fatal(err)
		}
		event.Props.Start = start
  }
	if eventFlags.Changed(eventFlag_End) { // End
		endString, _ := eventFlags.GetString(eventFlag_End)
		end, err := ian.ParseDateTime(endString, GetTimeZone())
		if err != nil {
			log.Fatal(err)
		}
		event.Props.End = end
  }
	if eventFlags.Changed(eventFlag_AllDay) { // All day
		event.Props.AllDay, _ = eventFlags.GetBool(eventFlag_AllDay)
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
  if eventFlags.Changed(eventFlag_Calendar) { // Calendar
    event.Path, _ = eventFlags.GetString(eventFlag_Calendar)
    if ian.IsPathInCache(event.Path) {
      log.Fatal("cannot set calendar to inside cache")
    }
    fmt.Println("note: event is being moved to another file location.")
  }
  if eventFlags.Changed(eventFlag_Recurrence) { // Recurrence
    recurrenceFlag, _ := eventFlags.GetString(eventFlag_Recurrence)
		rruleSet, err := rrule.StrToRRuleSet(recurrenceFlag)
		if err != nil {
			log.Fatal("'recurrence' set is invalid: ", err)
		}
		rruleSet.DTStart(event.Props.Start)
		event.Props.Rrule = rruleSet.String()
  }

  var modified []string
  for _, flag := range flags {
    if eventFlags.Changed(flag) {
      modified = append(modified, flag)
    }
  }

	if len(modified) == 0 {
		log.Fatal("no changes described. check the help page for a list of values to change.")
	}
	event.Props.Modified = time.Now().In(GetTimeZone())
	if err := event.Props.Validate(); err != nil {
		log.Fatalf("verification failed: %s", err)
	}
  checkCollision(instance, event.Props)

	event.Write(instance)
  fmt.Printf("'%s' has been updated; %s\n", event.Path, strings.Join(modified, ", "))

  instance.Sync(ian.SyncEvent{
    Type: ian.SyncEventUpdate,
    Files: event.GetRealPath(instance),
    Message: fmt.Sprintf("ian: edit event '%s'; %s", event.Path, strings.Join(modified, ", ")),
  }, false, nil)
}
