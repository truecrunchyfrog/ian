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
		start, err := icalEvent.DateTimeStart(time.UTC)
		if err != nil {
			return nil, err
		}
		end, err := icalEvent.DateTimeEnd(time.UTC)
		if err != nil {
			return nil, err
		}

		var allDay bool
		if start.Equal(end) || start.AddDate(0, 0, 1).Equal(end) {
			allDay = true
			end = start.AddDate(0, 0, 1).Add(-time.Second)
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
		created, _ := icalEvent.Props.DateTime(ical.PropCreated, time.UTC)
		modified, _ := icalEvent.Props.DateTime(ical.PropLastModified, time.UTC)

		props := EventProperties{
			Uid:         uid,
			Summary:     summary,
			Description: description,
			Location:    location,
			Url:         url,
			Start:       start,
			End:         end,
			AllDay:      allDay,
			Recurrence:  Recurrence{rrule, rdate, exdate},
			Created:     created,
			Modified:    modified,
		}
		eventsProps = append(eventsProps, props)
	}

	return eventsProps, nil
}

func ToIcal(events []Event) *ical.Calendar {
	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropProductID, "-//ian//ian calendar migration")
	now := time.Now()

	for _, event := range events {
		icalEvent := ical.NewEvent()

		icalEvent.Props.SetText(ical.PropUID, event.Props.Uid)

		icalEvent.Props.SetDateTime(ical.PropCreated, event.Props.Created)
		icalEvent.Props.SetDateTime(ical.PropLastModified, event.Props.Modified)

		icalEvent.Props.SetDateTime(ical.PropDateTimeStamp, now)

		icalEvent.Props.SetDateTime(ical.PropDateTimeStart, event.Props.Start)
		icalEvent.Props.SetDateTime(ical.PropDateTimeEnd, event.Props.End)

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

func SerializeIcal(ics *ical.Calendar) ([]byte, error) {
	var buf bytes.Buffer
	if err := ical.NewEncoder(&buf).Encode(ics); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
