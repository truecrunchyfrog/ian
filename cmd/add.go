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
	Long:    "The arguments 'start' and 'end' support many different formats. The recommended format for most events is dd/mm, or dd/mm hh:mm with time. A time zone can be appended with +-hhmm or 'UTC'-like format.",
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

	summary, _ := eventFlags.GetString(eventFlag_Summary)
	if summary == "" {
		log.Fatal("'summary' cannot be empty")
	}
	start, _ := eventFlags.GetString(eventFlag_Start)
	end, _ := eventFlags.GetString(eventFlag_End)
	description, _ := eventFlags.GetString(eventFlag_Description)
	location, _ := eventFlags.GetString(eventFlag_Location)
	url, _ := eventFlags.GetString(eventFlag_Url)
	calendar, _ := eventFlags.GetString(eventFlag_Calendar)

	duration, err := eventFlags.GetDuration(eventFlag_Duration)
	if err != nil {
		log.Fatal(err)
	}
	hours, _ := eventFlags.GetStringSlice(eventFlag_Hours)

	if (end != "" && (duration != 0 || len(hours) != 0)) ||
		(duration != 0 && (end != "" || len(hours) != 0)) {
		log.Fatal("'end', 'hours' and 'duration' are mutually exclusive")
	}

	startDate, err := ian.ParseDateTime(start, GetTimeZone())
	if err != nil {
		log.Fatal(err)
	}

	var endDate time.Time
	var allDay bool
	switch {
	case end != "":
		var err error
		endDate, err = ian.ParseDateTime(end, GetTimeZone())
		if err != nil {
			log.Fatal(err)
		}
	case duration != 0:
		endDate = startDate.Add(duration)
	case len(hours) != 0:
		if err := handleHours(hours, &startDate, &endDate); err != nil {
			log.Fatal(err)
		}
	default:
		if h, m, s := startDate.Clock(); h+m+s == 0 {
			endDate = startDate.AddDate(0, 0, 1).Add(-time.Second)
			allDay = true
		} else {
			endDate = startDate.Add(time.Hour)
		}
	}

  rrule, _ := eventFlags.GetString(eventFlag_Rrule)
  rdate, _ := eventFlags.GetString(eventFlag_Rdate)
  exdate, _ := eventFlags.GetString(eventFlag_ExDate)

	now := time.Now().In(GetTimeZone())

	props := ian.EventProperties{
		Uid:         ian.GenerateUid(),
		Summary:     summary,
		Description: description,
		Location:    location,
		Url:         url,
		Start:       startDate,
		End:         endDate,
		AllDay:      allDay,
		Recurrence:  ian.Recurrence{
			RRule:  rrule,
			RDate:  rdate,
			ExDate: exdate,
		},
		Created:     now,
		Modified:    now,
	}

	if err := props.Validate(); err != nil {
		log.Fatal("invalid event: ", err)
	}

	instance, err := ian.CreateInstance(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

	events, _, _ := instance.ReadEvents(props.GetTimeRange())

	checkCollision(&events, props)

	event, err := instance.CreateEvent(props, calendar)
	if err != nil {
		log.Fatal(err)
	}

	eventDuration := endDate.Sub(startDate)

	fmt.Printf("%s\n\n%s (%s)\n%s\n",
		props.Summary,
		props.Start.Format(ian.DefaultTimeLayout),
		ian.DurationToString(eventDuration),
		props.End.Format(ian.DefaultTimeLayout),
	)

	instance.Sync(ian.SyncEvent{
		Type:    ian.SyncEventCreate,
		Files:   event.GetRealPath(instance),
		Message: fmt.Sprintf("ian: create event '%s'", event.Path),
	}, false, nil)
}
