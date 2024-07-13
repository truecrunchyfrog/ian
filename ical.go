package ian

import (
	"fmt"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/teambition/rrule-go"
)

const icalTimeLayout string = "20060102T150405Z"

func parseIcalTime(input string) (time.Time, error) {
	return time.Parse(icalTimeLayout, input)
}

func getProp(icalEvent *ics.VEvent, prop ics.Property) string {
	ianaProp := icalEvent.GetProperty(ics.ComponentProperty(prop))
	if ianaProp == nil {
		return ""
	}
	return ianaProp.Value
}

func FromIcal(ical *ics.Calendar) ([]EventProperties, error) {
	eventsProps := []EventProperties{}

	icalEvents := ical.Events()
	for _, icalEvent := range icalEvents {
		start, err := icalEvent.GetStartAt()
		if err != nil {
			return nil, err
		}
		end, err := icalEvent.GetEndAt()
		if err != nil {
			return nil, err
		}

		var allDay bool
		if start.Equal(end) || start.AddDate(0, 0, 1).Equal(end) {
			allDay = true
			var err error
			if start, err = icalEvent.GetAllDayStartAt(); err != nil {
				return nil, err
			}

			end = start.AddDate(0, 0, 1).Add(-time.Second)
		}

		summary := getProp(icalEvent, ics.PropertySummary)
		description := getProp(icalEvent, ics.PropertyDescription)
		location := getProp(icalEvent, ics.PropertyLocation)
		url := getProp(icalEvent, ics.PropertyUrl)

		// Rrules:
		rruleSet := rrule.Set{}

		if rruleString := getProp(icalEvent, ics.PropertyRrule); rruleString != "" {
			rr, err := rrule.StrToRRule(rruleString)
			if err != nil {
				return nil, fmt.Errorf("RRule parse failed: %s", err)
			}
			rruleSet.RRule(rr)
			rruleSet.DTStart(start)
		}

		if rdateString := getProp(icalEvent, ics.PropertyRdate); rdateString != "" {
			rd, err := rrule.StrToDates(rdateString)
			if err != nil {
				return nil, fmt.Errorf("RDate parse failed: %s", err)
			}
			for _, d := range rd {
				rruleSet.RDate(d)
			}
		}

		if exdateString := getProp(icalEvent, ics.PropertyExdate); exdateString != "" {
			xd, err := rrule.StrToDates(exdateString)
			if err != nil {
				return nil, fmt.Errorf("ExDate parse failed: %s", err)
			}
			for _, d := range xd {
				rruleSet.ExDate(d)
			}
		}

		// Ignore errors for these, since they may not exist.
		created, _ := parseIcalTime(getProp(icalEvent, ics.PropertyCreated))
		modified, _ := parseIcalTime(getProp(icalEvent, ics.PropertyLastModified))

		props := EventProperties{
			Summary:     summary,
			Description: description,
			Location:    location,
			Url:         url,
			Start:       start,
			End:         end,
			AllDay:      allDay,
			Rrule:       rruleSet.String(),
			Created:     created,
			Modified:    modified,
		}
		eventsProps = append(eventsProps, props)
	}

	return eventsProps, nil
}
