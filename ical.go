package ian

import (
	"time"

	"github.com/arran4/golang-ical"
)

const icalTimeLayout string = "20060102T150405Z"

func parseIcalTime(input string) (time.Time, error) {
  return time.Parse(icalTimeLayout, input)
}

func getProp(icalEvent *ics.VEvent, prop ics.Property) string {
  return icalEvent.GetProperty(ics.ComponentProperty(prop)).Value
}

func FromIcal(ical *ics.Calendar) ([]CalendarEvent, error) {
	events := []CalendarEvent{}

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

		summary := getProp(icalEvent, ics.PropertySummary)
    description := getProp(icalEvent, ics.PropertyDescription)
		location := getProp(icalEvent, ics.PropertyLocation)
    url := getProp(icalEvent, ics.PropertyUrl)

    // Ignore errors for these, since they may not exist.
    created, _ := parseIcalTime(getProp(icalEvent, ics.PropertyCreated))
    modified, _ := parseIcalTime(getProp(icalEvent, ics.PropertyLastModified))

		// TODO add more properties: Rrule, etc.

		event := CalendarEvent{
			Summary:     summary,
			Description: description,
			Location:    location,
			Url:         url,
			Start:       start,
			End:         end,
			Created:     created,
			Modified:    modified,
			Constant:    true, // It's a static calendar
		}
		events = append(events, event)
	}

	return nil, nil
}
