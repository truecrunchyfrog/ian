package ian

import (
	"time"

	ics "github.com/arran4/golang-ical"
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
		if start.Equal(end) {
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

		// Ignore errors for these, since they may not exist.
		created, _ := parseIcalTime(getProp(icalEvent, ics.PropertyCreated))
		modified, _ := parseIcalTime(getProp(icalEvent, ics.PropertyLastModified))

		// TODO add more properties: Rrule, etc.

		props := EventProperties{
			Summary:     summary,
			Description: description,
			Location:    location,
			Url:         url,
			Start:       start,
			End:         end,
			AllDay:      allDay,
			Created:     created,
			Modified:    modified,
		}
		eventsProps = append(eventsProps, props)
	}

	return eventsProps, nil
}
