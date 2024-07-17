package ian

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/spf13/viper"
	"github.com/teambition/rrule-go"
)

type EventType int

const (
	EventTypeNormal EventType = 1 << iota
	EventTypeCache
	EventTypeRecurrence
)

type Event struct {
	// Path to event file relative to root.
	// Use `filepath.ToSlash(filepath.Rel(root, filename))`.
	Path  string
	Props EventProperties
	Type  EventType
	// Constant is true if the event should not be changed. Used for source events (cache) or the event is generated from a recurrance (RRule).
	Constant bool
	// Parent is the parent event if this event is generated from a recurrence rule. Otherwise nil.
	Parent *Event
}

func (event *Event) GetFilepath(instance *Instance) string {
	return filepath.Join(instance.Root, filepath.FromSlash(event.Path))
}

func (event *Event) GetCalendarName() string {
	return SanitizePath(path.Dir(event.Path))
}

// Write writes the event to the appropriate location in 'instance'.
// Creates any necessary directories.
func (event *Event) Write(instance *Instance) error {
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(event.Props); err != nil {
		return err
	}

	path := event.GetFilepath(instance)
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
	Uid string

	Summary     string
	Description string
	Location    string
	Url         string

	// Start is an inclusive datetime representing when the event begins.
	Start time.Time
	// End is a non-inclusive datetime representing when the event ends.
	End time.Time

	Recurrence Recurrence

	Created  time.Time
	Modified time.Time
}

type Recurrence struct {
	RRule  string
	RDate  string
	ExDate string
}

func (rec *Recurrence) IsThereRecurrence() bool {
	return rec.RRule != "" || rec.RDate != ""
}

func (props *EventProperties) GetRruleSet() (rrule.Set, error) {
	set := rrule.Set{}

	if s := props.Recurrence.RRule; s != "" {
		rr, err := rrule.StrToRRule(s)
		if err != nil {
			return rrule.Set{}, fmt.Errorf("RRULE parse failed: %s", err)
		}
		set.RRule(rr)
		set.DTStart(props.Start)
	}

	if s := props.Recurrence.RDate; s != "" {
		rd, err := rrule.StrToDates(s)
		if err != nil {
			return rrule.Set{}, fmt.Errorf("RDATE parse failed: %s", err)
		}
		set.SetRDates(rd)
	}

	if s := props.Recurrence.ExDate; s != "" {
		xd, err := rrule.StrToDates(s)
		if err != nil {
			return rrule.Set{}, fmt.Errorf("EXDATE parse failed: %s", err)
		}
		set.SetExDates(xd)
	}

	return set, nil
}

func (props *EventProperties) IsAllDay() bool {
	h, m, s := props.Start.Clock()
	h2, m2, s2 := props.End.Clock()
	return h == 0 && m == 0 && s == 0 && h2 == 0 && m2 == 0 && s2 == 0
}

func (props *EventProperties) GetTimeRange() TimeRange {
	return TimeRange{
		From: props.Start,
		To:   props.End,
	}
}

func (p *EventProperties) Validate() error {
	if viper.GetBool("no-validation") {
		return nil
	}
	switch {
	case p.Uid == "":
		return errors.New("uid cannot be empty")
	case p.Summary == "":
		return errors.New("summary cannot be empty")
	case p.Start.After(p.End):
		return errors.New("start cannot be chronologically after end")
	case p.Created.After(p.Modified):
		return errors.New("created cannot be chronologically after modified")
	}

	return nil
}

func (props *EventProperties) FormatName() string {
	return strings.NewReplacer(
		"/", "-",
		`\`, "-",
		".", "_",
	).Replace(props.Summary)
}

func GetEvent(events *[]Event, path string) (*Event, error) {
	for _, ev := range *events {
		if ev.Path == path {
			return &ev, nil
		}
	}
	return nil, fmt.Errorf("no such event '%s'", path)
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
