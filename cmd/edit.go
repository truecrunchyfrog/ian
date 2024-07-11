package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

const editFlag_Summary = "summary"
const editFlag_Start = "start"
const editFlag_End = "end"
const editFlag_AllDay = "all-day"
const editFlag_Description = "description"
const editFlag_Location = "location"
const editFlag_Url = "url"

func init() {
	editCmd.Flags().String(editFlag_Summary, "", "Event brief.")

	editCmd.Flags().String(editFlag_Start, "", "Start date.")
	editCmd.Flags().String(editFlag_End, "", "End date.")
	editCmd.Flags().Bool(editFlag_AllDay, false, "If the event should be marked as all-day.")

	editCmd.Flags().String(editFlag_Description, "", "Detailed event description.")
	editCmd.Flags().String(editFlag_Location, "", "Where the event is taking place.")
	editCmd.Flags().String(editFlag_Url, "", "Online reference.")

	rootCmd.AddCommand(editCmd)
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

	modified := true

	switch {
	case cmd.Flags().Lookup(editFlag_Summary).Changed: // Summary
		event.Props.Summary, _ = cmd.Flags().GetString(editFlag_Summary)
	case cmd.Flags().Lookup(editFlag_Start).Changed: // Start
		startString, _ := cmd.Flags().GetString(editFlag_Start)
		start, err := ian.ParseDateTime(startString, GetTimeZone())
		if err != nil {
			log.Fatal(err)
		}
		event.Props.Start = start
	case cmd.Flags().Lookup(editFlag_End).Changed: // End
		endString, _ := cmd.Flags().GetString(editFlag_End)
		end, err := ian.ParseDateTime(endString, GetTimeZone())
		if err != nil {
			log.Fatal(err)
		}
		event.Props.End = end
	case cmd.Flags().Lookup(editFlag_AllDay).Changed: // All day
		event.Props.AllDay, _ = cmd.Flags().GetBool(editFlag_AllDay)

	case cmd.Flags().Lookup(editFlag_Description).Changed: // Description
		event.Props.Description, _ = cmd.Flags().GetString(editFlag_Description)
	case cmd.Flags().Lookup(editFlag_Location).Changed: // Location
		event.Props.Location, _ = cmd.Flags().GetString(editFlag_Location)
	case cmd.Flags().Lookup(editFlag_Url).Changed: // URL
		event.Props.Url, _ = cmd.Flags().GetString(editFlag_Url)
	default:
		modified = false
	}

	if !modified {
		log.Fatal("no changes described. use the flags listed below to modify the event.")
		cmd.Help()
	}
	event.Props.Modified = time.Now().In(GetTimeZone())
	if err := event.Props.Validate(); err != nil {
		log.Fatalf("verification failed: %s", err)
	}
  checkCollision(instance, event.Props)

	event.Write(instance)
	fmt.Printf("'%s' has been updated\n", event.Path)

  instance.Sync(ian.SyncEvent{
    Type: ian.SyncEventUpdate,
    Files: event.GetRealPath(instance),
    Message: fmt.Sprintf("ian: edit event '%s'", event.Path),
  }, false, nil)
}
