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

// EventPath specifies where to find an event, and its calendar.
type EventPath interface {
	Calendar() string
	Name() string

	String() string
	Filepath(*Instance) string
}

type eventPath struct {
	calendar,
	name string
}

// TODO implement check against using '.'?
// and when its needed (for cache (calendar) and recurrence (event name)), create the path without proper construction (raw)
// NewEventPath safely constructs an EventPath from a calendar and name, which can be easily used for determining paths and filenames.
// name may be modified.
func NewEventPath(calendar, name string) (EventPath, error) {
	illegalChars := "\n/" + string(filepath.Separator)

	if calendar == "" || strings.ContainsAny(calendar, illegalChars + " ") {
		return nil, fmt.Errorf("calendar '%s' contains illegal characters", calendar)
	}

  name = strings.TrimSpace(name)
	if name == "" || strings.ContainsAny(name, illegalChars) {
		return nil, fmt.Errorf("name '%s' contains illegal characters", name)
	}

	return &eventPath{
		calendar: calendar,
		name:     name,
	}, nil
}

// NewFreeEventPath is like NewEventPath, but ensures that the filename is available, possibly by changing it.
func NewFreeEventPath(instance *Instance, calendar, name string) (EventPath, error) {
  safeName, err := instance.getAvailableFilename(filepath.Join(instance.Root, calendar), name)
  if err != nil {
    return nil, err
  }
  return NewEventPath(calendar, safeName)
}

func ParseEventPath(input string) (EventPath, error) {
	cal, name := path.Split(input)
	return NewEventPath(cal, name)
}

func (p *eventPath) Calendar() string {
	return p.calendar
}

func (p *eventPath) Name() string {
	return p.name
}

func (p *eventPath) String() string {
	return path.Join(p.calendar, p.name)
}

func (p *eventPath) Filepath(instance *Instance) string {
	return filepath.Join(instance.Root, p.calendar, p.name)
}

type Event struct {
	Path EventPath

	Props EventProperties
	Type  EventType
	// Constant is true if the event should not be changed. Used for source events (cache) or the event is generated from a recurrance (RRule).
	Constant bool
	// Parent is the parent event if this event is generated from a recurrence rule. Otherwise nil.
	Parent *Event
}

// Write writes the event to the appropriate location in 'instance'.
func (event *Event) Write(instance *Instance) error {
  return event.Props.Write(event.Path.Filepath(instance))
}

func (event *Event) String() string {
	return event.Path.String()
}

type Recurrence struct {
	RRule  string
	RDate  string
	ExDate string
}

func (rec *Recurrence) IsThereRecurrence() bool {
	return rec.RRule != "" || rec.RDate != ""
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

func (props *EventProperties) Write(file string) error {
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(props); err != nil {
		return err
	}

	CreateDir(filepath.Dir(file)) // Create parent folder(s) leading to path.

	if err := os.WriteFile(file, buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
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

// TODO change path to type EventPath?
func GetEvent(events *[]Event, path string) (*Event, error) {
	for _, ev := range *events {
		if ev.Path.String() == path {
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

func BuildEvent(path EventPath, props EventProperties, eType EventType) (Event, error) {
	if err := props.Validate(); err != nil {
		return Event{}, fmt.Errorf("failed validation: %s", err)
	}

	return Event{
		Path:     path,
		Props:    props,
		Type:     eType,
		Constant: eType == EventTypeCache,
	}, nil
}
