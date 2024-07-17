package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

func init() {
	eventPropsCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:     "add [summary] [start] [end]",
	Aliases: []string{"a", "create", "cr", "make", "mk", "new", "n"},
	Short:   "Create a new event",
	Args:    cobra.RangeArgs(0, 3),
	Run:     addCmdRun,
}

func addCmdRun(cmd *cobra.Command, args []string) {
	if len(args) >= 1 {
		eventFlags.Set(eventFlag_Summary, args[0])
	}

	if len(args) >= 2 {
		eventFlags.Set(eventFlag_Start, args[1])
	}

	if len(args) >= 3 {
		eventFlags.Set(eventFlag_End, args[2])
	}

	var props ian.EventProperties

	props.Summary, _ = eventFlags.GetString(eventFlag_Summary)
	if props.Summary == "" {
		log.Fatal("'summary' cannot be empty")
	}
	props.Description, _ = eventFlags.GetString(eventFlag_Description)
	props.Location, _ = eventFlags.GetString(eventFlag_Location)
	props.Url, _ = eventFlags.GetString(eventFlag_Url)

	start, _ := eventFlags.GetString(eventFlag_Start)
	end, _ := eventFlags.GetString(eventFlag_End)

	duration, err := eventFlags.GetDuration(eventFlag_Duration)
	if err != nil {
		log.Fatal(err)
	}
	hours, _ := eventFlags.GetStringSlice(eventFlag_Hours)

	if (end != "" && (duration != 0 || len(hours) != 0)) ||
		(duration != 0 && (end != "" || len(hours) != 0)) {
		log.Fatal("'end', 'hours' and 'duration' are mutually exclusive")
	}

	props.Start, err = ian.ParseDateTime(start, ian.GetTimeZone())
	if err != nil {
		log.Fatal(err)
	}

	switch {
	case end != "":
		var err error
		props.End, err = ian.ParseDateTime(end, ian.GetTimeZone())
		if err != nil {
			log.Fatal(err)
		}
	case duration != 0:
		props.End = props.Start.Add(duration)
	case len(hours) != 0:
		if err := handleHours(hours, &props.Start, &props.End); err != nil {
			log.Fatal(err)
		}
	default:
		// No end date provided.
		if h, m, s := props.Start.Clock(); h+m+s == 0 {
			// Start date had no time, so count it as the full day.
			props.End = props.Start.AddDate(0, 0, 1)
		} else {
			// Start date did have time, so add 1 hour.
			props.End = props.Start.Add(time.Hour)
		}
	}

	props.Recurrence.RRule, _ = eventFlags.GetString(eventFlag_Rrule)
	props.Recurrence.RDate, _ = eventFlags.GetString(eventFlag_Rdate)
	props.Recurrence.ExDate, _ = eventFlags.GetString(eventFlag_ExDate)

	props.Uid = ian.GenerateUid()

	now := time.Now().In(ian.GetTimeZone())
	props.Created = now
	props.Modified = now

	if err := props.Validate(); err != nil {
		log.Fatal("invalid event: ", err)
	}

	instance, err := ian.CreateInstance(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

	events, _, err := instance.ReadEvents(props.GetTimeRange())
  if err != nil {
    log.Fatal(err)
  }

	checkCollision(&events, props)

	calendar, _ := eventFlags.GetString(eventFlag_Calendar)
	event, err := instance.NewEvent(props, calendar)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n\n%s (%s)\n%s\n",
		props.Summary,
		props.Start.Format(ian.DefaultTimeLayout),
		ian.DurationToString(props.End.Sub(props.Start)),
		props.End.Format(ian.DefaultTimeLayout),
	)

	instance.Sync(func() error {
		return event.Write(instance)
	}, ian.SyncEvent{
		Type:    ian.SyncEventCreate,
		Files:   []string{event.GetFilepath(instance)},
		Message: fmt.Sprintf("ian: create event '%s'", event.Path),
	}, false, nil)
}
