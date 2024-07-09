package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

var calendar string

var description string
var location string
var url string
var duration time.Duration
var hours []string

func init() {
	addCmd.Flags().StringVarP(&calendar, "calendar", "c", "", "Add the event to a specified calendar (e.g. 'share'), instead of the main calendar (''). Calendars are the directories inside the root. Calendars can be recursive, e.g.: 'share/family'.")

	addCmd.Flags().StringVarP(&description, "description", "e", "", "A more detailed event description.")
	addCmd.Flags().StringVarP(&location, "location", "l", "", "Where the event will take place (i.e. address).")
	addCmd.Flags().StringVarP(&url, "url", "u", "", "A URL relevant to the event.")
	addCmd.Flags().DurationVarP(&duration, "duration", "d", 0, "Duration of the event from start to end. Use instead of providing the 'end' argument. Example: --duration 1h30m to set 'end' to 'start' with 1 hour and 30 minutes added.")
	addCmd.Flags().StringSliceVarP(&hours, "hours", "H", nil, "Time of the day(s). E.g.: '-h 09:00,17:00' to complement the start date with the time '09:00' and set the 'end' date to the same day but with the time '17:00', or '-h 22:00,05:00' to complement the start date with the time '22:00' and set the 'end' date to the day after with the time '05:00'. The second time parameter can be replaced with a duration, like in --duration.")
	addCmd.MarkFlagsMutuallyExclusive("duration", "hours")

	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add <summary> <start> [end]",
	Short: "Create a new event",
	Long:  "The arguments 'start' and 'end' support many different formats. The recommended format for most events is dd/mm, or dd/mm hh:mm with time. A time zone can be appended with +-hhmm or 'UTC'-like format.",
	Args:  cobra.RangeArgs(2, 3),
	Run:   addCmdRun,
}

func addCmdRun(cmd *cobra.Command, args []string) {
	if len(args) >= 3 && (duration != 0 || hours != nil) {
		log.Fatal("'end' and '--duration' cannot be combined")
	}

	startDate, err := ian.ParseDateTime(args[1], GetTimeZone())
	if err != nil {
		log.Fatal(err)
	}

	var endDate time.Time
	var allDay bool
	switch {
	case len(args) >= 3:
		var err error
		endDate, err = ian.ParseDateTime(args[2], GetTimeZone())
		if err != nil {
			log.Fatal(err)
		}
	case duration != 0:
		endDate = startDate.Add(duration)
	case hours != nil:
		if len(hours) != 2 {
			log.Fatal("--hours must have exactly two parameters, like: '--hours 09:00,17:00'.")
		}
		// Complement start date with the first time parameter.
		startT, err := ian.ParseTimeOnly(hours[0])
		if err != nil {
			log.Fatal(err)
		}
		startDate = time.Date(
			startDate.Year(),
			startDate.Month(),
			startDate.Day(),
			startT.Hour(),       // Replace time
			startT.Minute(),     //
			startT.Second(),     //
			startT.Nanosecond(), //
			startDate.Location(),
		)

		// Create end date.
		endT, err := ian.ParseTimeOnly(hours[1])
		if err == nil {
			dayOffset := 0
			if endT.Before(startT) {
				// Time is less, so it's the day after.
				dayOffset = 1
			}
			endDate = time.Date(
				startDate.Year(),
				startDate.Month(),
				startDate.Day()+dayOffset,
				endT.Hour(),
				endT.Minute(),
				endT.Second(),
				endT.Nanosecond(),
				startDate.Location(),
			)
		} else {
			// Maybe it's a duration instead.
			d, durErr := time.ParseDuration(hours[1])
			if durErr != nil {
				log.Fatalf("the second parameter in --hours: '%s', can neither be parsed as a time ('%s'), or duration ('%s').\n", hours[1], err, durErr)
			}
			if d < 0 {
				log.Fatal("--hours duration cannot be negative")
			}
			// It's a duration!
			endDate = startDate.Add(d)
		}
	default:
		if h, m, s := startDate.Clock(); h+m+s == 0 {
      endDate = startDate.AddDate(0, 0, 1).Add(-time.Second)
      allDay = true
		} else {
      endDate = startDate.Add(time.Hour)
    }
	}

	now := time.Now()

	props := ian.EventProperties{
		Summary:     args[0],
		Description: description,
		Location:    location,
		Url:         url,
		Start:       startDate,
		End:         endDate,
		AllDay:      allDay,
		Created:     now,
		Modified:    now,
	}

	if err := props.Verify(); err != nil {
		log.Fatal("invalid event: ", err)
	}

	instance, err := ian.CreateInstance(GetRoot(), true)
	if err != nil {
		log.Fatal(err)
	}

  checkCollision(instance, props)

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
    Type: ian.SyncEventCreate,
    Files: event.GetRealPath(instance),
    Message: fmt.Sprintf("ian: create event '%s'", event.Path),
  })
}
