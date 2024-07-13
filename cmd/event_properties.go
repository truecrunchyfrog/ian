package cmd

import (
	"errors"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

const eventFlag_Calendar = "calendar"
const eventFlag_Summary = "summary"
const eventFlag_Start = "start"
const eventFlag_End = "end"
const eventFlag_AllDay = "all-day"
const eventFlag_Description = "description"
const eventFlag_Location = "location"
const eventFlag_Url = "url"
const eventFlag_Duration = "duration"
const eventFlag_Hours = "hours"

const eventFlag_Rrule = "rrule"
const eventFlag_Rdate = "rdate"
const eventFlag_ExDate = "exdate"

var eventPropsCmd = &cobra.Command{
	Use:     "event",
	Aliases: []string{"ev"},
}
var eventFlags = eventPropsCmd.PersistentFlags()

func init() {
	eventFlags.StringP(eventFlag_Calendar, "c", "", "Specify calendar to place event in.")
	eventFlags.StringP(eventFlag_Summary, "S", "", "Event brief.")

	eventFlags.StringP(eventFlag_Start, "s", "", "Start date.")
	eventFlags.StringP(eventFlag_End, "e", "", "End date.")
	eventFlags.BoolP(eventFlag_AllDay, "a", false, "If the event should be marked as all-day.")
	eventFlags.String(eventFlag_Rrule, "", "An RRULE expression according to iCalendar RFC 5545.")
	eventFlags.String(eventFlag_Rdate, "", "An RDATE expression according to iCalendar RFC 5545.")
	eventFlags.String(eventFlag_ExDate, "", "An EXDATE expression according to iCalendar RFC 5545.")

	eventFlags.StringP(eventFlag_Description, "D", "", "Detailed event description.")
	eventFlags.StringP(eventFlag_Location, "l", "", "Where the event is taking place (e.g. address).")
	eventFlags.StringP(eventFlag_Url, "u", "", "A URL relevant to the event.")

	eventFlags.DurationP(eventFlag_Duration, "d", 0, "Duration of the event from start to end. Use instead of providing the 'end' argument. Example: '--duration 1h30m' to set 'end' to 'start' with 1 hour and 30 minutes added.")
	eventFlags.StringSliceP(eventFlag_Hours, "H", nil, "Time of the day(s). E.g.: '-h 09:00,17:00' to complement the start date with the time '09:00' and set the 'end' date to the same day but with the time '17:00', or '-h 22:00,05:00' to complement the start date with the time '22:00' and set the 'end' date to the day after with the time '05:00'. The second time parameter can be replaced with a duration, like in --duration.")

	eventPropsCmd.MarkFlagsMutuallyExclusive(eventFlag_End, eventFlag_Duration)
	eventPropsCmd.MarkFlagsMutuallyExclusive(eventFlag_End, eventFlag_Hours)
	eventPropsCmd.MarkFlagsMutuallyExclusive(eventFlag_Duration, eventFlag_Hours)

	rootCmd.AddCommand(eventPropsCmd)
}

func handleHours(hours []string, startDate *time.Time, endDate *time.Time) error {
	if len(hours) != 2 {
		return errors.New("'hours' must have exactly two parameters, like: '--hours 09:00,17:00'.")
	}

	// Complement start date with the first time parameter.
	startT, err := ian.ParseTimeOnly(hours[0])
	if err != nil {
		return err
	}
	*startDate = time.Date(
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
		*endDate = time.Date(
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
			log.Fatalf("the second parameter in 'hours': '%s', can neither be parsed as a time ('%s'), or duration ('%s').\n", hours[1], err, durErr)
		}
		if d < 0 {
			return errors.New("'hours' duration cannot be negative")
		}
		// It's a duration!
		*endDate = startDate.Add(d)
	}

	return nil
}
