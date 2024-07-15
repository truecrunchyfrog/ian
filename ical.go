package ian

import (
	"bytes"
	"io"
	"time"

	"github.com/emersion/go-ical"
)

func FromIcal(cal *ical.Calendar) ([]EventProperties, error) {
	eventsProps := []EventProperties{}

	for _, icalEvent := range cal.Events() {
		start, err := icalEvent.DateTimeStart(GetTimeZone())
		if err != nil {
			return nil, err
		}
		end, err := icalEvent.DateTimeEnd(GetTimeZone())
		if err != nil {
			return nil, err
		}

		if h, m, s := start.Clock(); h+m+s == 0 && start.Equal(end) {
			// Event is an *alternatively* formatted all-day event.
			end = start.AddDate(0, 0, 1)
		}

		var uid, summary, description, location, url, rrule, rdate, exdate string

		textPropsMap := map[*string]string{
			&uid:         ical.PropUID,
			&summary:     ical.PropSummary,
			&description: ical.PropDescription,
			&location:    ical.PropLocation,
			&url:         ical.PropURL,
			&rrule:       ical.PropRecurrenceRule,
			&rdate:       ical.PropRecurrenceDates,
			&exdate:      ical.PropExceptionDates,
		}

		for dest, propName := range textPropsMap {
			prop := icalEvent.Props.Get(propName)
			if prop != nil {
				*dest = prop.Value
			}
		}

		// Ignore errors for these, since they may not exist.
		created, _ := icalEvent.Props.DateTime(ical.PropCreated, GetTimeZone())
		modified, _ := icalEvent.Props.DateTime(ical.PropLastModified, GetTimeZone())

		props := EventProperties{
			Uid:         uid,
			Summary:     summary,
			Description: description,
			Location:    location,
			Url:         url,
			Start:       start,
			End:         end,
			Recurrence:  Recurrence{rrule, rdate, exdate},
			Created:     created,
			Modified:    modified,
		}
		eventsProps = append(eventsProps, props)
	}

	return eventsProps, nil
}

func ToIcal(events []Event, calendarName string) *ical.Calendar {
	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//ian//ian calendar migration")

  if calendarName != "" {
    cal.Props.SetText("X-WR-NAME", calendarName)
  }

	now := time.Now()

	for _, event := range events {
		icalEvent := ical.NewEvent()

		icalEvent.Props.SetText(ical.PropUID, event.Props.Uid)

		icalEvent.Props.SetDateTime(ical.PropCreated, event.Props.Created)
		icalEvent.Props.SetDateTime(ical.PropLastModified, event.Props.Modified)

		icalEvent.Props.SetDateTime(ical.PropDateTimeStamp, now)

		if !event.Props.IsAllDay() {
			icalEvent.Props.SetDateTime(ical.PropDateTimeStart, event.Props.Start)
			icalEvent.Props.SetDateTime(ical.PropDateTimeEnd, event.Props.End)
		} else {
			icalEvent.Props.SetDate(ical.PropDateTimeStart, event.Props.Start)
			icalEvent.Props.SetDate(ical.PropDateTimeEnd, event.Props.Start.AddDate(0, 0, 1))
		}

		icalEvent.Props.SetText(ical.PropSummary, event.Props.Summary)

		optionalProps := map[string]string{
			event.Props.Description:       ical.PropDescription,
			event.Props.Location:          ical.PropLocation,
			event.Props.Url:               ical.PropURL,
			event.Props.Recurrence.RRule:  ical.PropRecurrenceRule,
			event.Props.Recurrence.RDate:  ical.PropRecurrenceDates,
			event.Props.Recurrence.ExDate: ical.PropExceptionDates,
		}

		for value, assignAs := range optionalProps {
			if value != "" {
				icalEvent.Props.SetText(assignAs, value)
			}
		}

		cal.Children = append(cal.Children, icalEvent.Component)
	}

	return cal
}

func ParseIcal(r io.Reader) (*ical.Calendar, error) {
	ics, err := ical.NewDecoder(r).Decode()
	if err != nil {
		return nil, err
	}
	return ics, nil
}

func SerializeIcal(ics *ical.Calendar) (bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := ical.NewEncoder(&buf).Encode(ics); err != nil {
		return bytes.Buffer{}, err
	}
	return buf, nil
}
