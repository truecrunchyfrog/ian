package ian

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Event struct {
	// Path to event file relative to root.
	// Use `filepath.Rel(root, filename)`.
	Path  string // TODO make path the same on all platforms (filepath.ToSlash()/FromSlash())
	Props EventProperties

	// Constant is true if the event should not be changed. Used for source events (cache) or the event is generated from a recurrance (RRule).
	Constant bool

	// Parent is the parent event if this event is generated from a recurrence rule. Otherwise nil.
	Parent *Event
}

func (event *Event) GetRealPath(instance *Instance) string {
	return filepath.Join(instance.Root, event.Path)
}

func (event *Event) GetCalendarName() string {
	return filepath.Dir(event.Path)
}

// Write writes the event to the appropriate location in 'instance'.
// Creates any necessary directories.
func (event *Event) Write(instance *Instance) error {
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(event.Props); err != nil {
		return err
	}

	path := event.GetRealPath(instance)
	CreateDir(filepath.Dir(path)) // Create parent folder(s) leading to path.

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func (event *Event) String() string {
	return event.Path
}

type EventProperties struct {
	Summary     string
	Description string
	Location    string
	Url         string

	Start  time.Time
	End    time.Time
	AllDay bool

	Rrule string

	Created  time.Time
	Modified time.Time
}

func (props *EventProperties) GetTimeRange() TimeRange {
	return TimeRange{
		From: props.Start,
		To:   props.End,
	}
}

func (p *EventProperties) Validate() error {
	switch {
	case p.Summary == "":
		return errors.New("summary cannot be empty")
	case p.Start.After(p.End):
		return errors.New("start cannot be chronologically after end")
	case p.Created.After(p.Modified):
		return errors.New("created cannot be chronologically after modified")
	case p.AllDay:
		if !p.Start.Equal(time.Date(p.Start.Year(), p.Start.Month(), p.Start.Day(), 0, 0, 0, 0, p.Start.Location())) {
			return errors.New("all-day event start must be at midnight")
		}
		if !p.Start.AddDate(0, 0, 1).Add(-time.Second).Equal(p.End) {
			return errors.New("all-day event end must be exactly 24 hours minus 1 second after start")
		}
	}

	return nil
}

func (props *EventProperties) FormatName() string {
	name := props.Summary
	//name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, `\`, "-")
	//name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, ".", "_")

	return name
}

func GetEvent(events *[]Event, relPath string) (*Event, error) {
	for _, ev := range *events {
		if ev.Path == relPath {
			return &ev, nil
		}
	}
	return nil, fmt.Errorf("no such event '%s'", relPath)
}

func FilterEvents(events *[]Event, filter func(*Event) bool) []Event {
	filtered := []Event{}

	for _, event := range *events {
		if filter(&event) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

func QueryEvents(events *[]Event, query string) []Event {
	return FilterEvents(events, func(e *Event) bool {
		return strings.Contains(e.Path, query) || strings.Contains(e.Props.Summary, query) || strings.Contains(e.Props.Description, query)
	})
}