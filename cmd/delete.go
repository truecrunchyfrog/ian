package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

func init() {
	eventPropsCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:     "rm event...",
	Aliases: []string{"remove", "rem", "delete", "del", "d"},
	Short:   "Delete event(s)",
	Run:     deleteCmdRun,
}

func deleteCmdRun(cmd *cobra.Command, args []string) {
	instance, err := ian.CreateInstance(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

	events, _, err := instance.ReadEvents(ian.TimeRange{})
	if err != nil {
		log.Fatal(err)
	}

	deleteEvents := []*ian.Event{}

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
		if event.Constant {
			log.Fatalf("'%s' is a constant event and cannot be deleted by itself.\n", event.Path)
		}
		if event.Props.Recurrence.IsThereRecurrence() {
			log.Printf("warning: '%s' is a recurring event and all recurrences will be deleted too.\n", event.Path)
		}
		deleteEvents = append(deleteEvents, event)
	}

	if len(deleteEvents) == 0 {
		log.Fatal("no event to delete")
	}

	syncMsg := "ian: delete "
	if len(deleteEvents) > 1 {
		syncMsg += fmt.Sprintf("%d events; ", len(deleteEvents))
	} else {
		syncMsg += "event: "
	}

	filesToDelete := []string{}

	for i, deleteEvent := range deleteEvents {
		filesToDelete = append(filesToDelete, deleteEvent.GetFilepath(instance))

		if i != 0 {
			syncMsg += ", "
		}
		syncMsg += "'" + deleteEvent.Path + "'"

		fmt.Printf("deleted event '%s'\n", deleteEvent.Path)
	}

	err = instance.Sync(func() error {
		for _, file := range filesToDelete {
			if err := os.Remove(file); err != nil {
				return err
			}
		}
		return nil
	}, ian.SyncEvent{
		Type:    ian.SyncEventDelete,
		Files:   filesToDelete,
		Message: syncMsg,
	}, false, nil)

	if err != nil {
		log.Fatal(err)
	}
}
