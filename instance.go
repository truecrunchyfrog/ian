package ian

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Instance struct {
	Root   string
	Config Config
}

// Work performs maintenance work and is run on every instance creation.
// It is used to e.g. update sources.
func (instance *Instance) Work() error {
	if err := instance.UpdateSources(); err != nil {
		return err
	}
	return nil
}

func (instance *Instance) clearDir(name string) error {
	return os.RemoveAll(filepath.Join(instance.Root, SanitizeFilepath(name)))
}

// NewEvent constructs a standard event based on properties, as a part of calendar.
// NewEvent does not write anything.
func (instance *Instance) NewEvent(props EventProperties, calendar string) (Event, error) {
	p, err := NewFreeEventPath(instance, calendar, props.FormatName())
	if err != nil {
		return Event{}, err
	}

	return Event{
		Path:  p,
		Props: props,
	}, nil
}

// WriteNewEvent creates an event in the instance by writing it.
func (instance *Instance) WriteNewEvent(props EventProperties, calendar string) (*Event, error) {
	event, err := instance.NewEvent(props, calendar)
	if err != nil {
		return nil, err
	}

	if err := event.Write(instance); err != nil {
		return nil, err
	}

	return &event, nil
}

// getAvailableFilename tries to generate an available name like originalName (with possible number suffix), in the directory dir.
//
// Note: Only the name is returned, NOT the entire path with dir.
func (instance *Instance) getAvailableFilename(dir, originalName string) (string, error) {
	var pathSuffix string

	for i := 2; ; i++ {
		if i > 50 {
			return "", errors.New("cannot create file with that name: tried to add numerical suffix up to 50, but files by those names already exist.")
		}

		if _, err := os.Stat(filepath.Join(dir, originalName+pathSuffix)); err != nil && os.IsNotExist(err) {
			// File does not exist: safe to write file
			return originalName + pathSuffix, nil
		} else {
			// File already exists
			pathSuffix = "_" + fmt.Sprint(i)
			continue
		}
	}
}

// ReadEvents reads all events in the instance that appear during the time range, and parses their recurrences.
// If the time range is empty (From.IsZero() && To.IsZero()), then all events are shown,
// and recurrences are shown within the range of the normal events.
func (instance *Instance) ReadEvents(timeRange TimeRange) ([]Event, []*Event, error) {
	events := []Event{}

	calDirs, err := os.ReadDir(instance.Root)
	if err != nil {
		return nil, nil, err
	}
	for _, calDir := range calDirs {
		if strings.HasPrefix(calDir.Name(), ".") {
			continue
		}
		if !calDir.IsDir() {
			log.Printf("warning: ignoring file '%s'. the root directory should only contain calendars (directories). any other files/directories should be prefixed with a dot ('.').\n", filepath.Join(instance.Root, calDir.Name()))
			continue
		}
		propsList, err := instance.readDir(filepath.Join(instance.Root, calDir.Name()))
		if err != nil {
			return nil, nil, err
		}

		for name, props := range propsList {
			path, err := NewEventPath(calDir.Name(), name)
			if err != nil {
				return nil, nil, err
			}
			events = append(events, Event{
				Path:  path,
				Props: props,
				Type:  EventTypeNormal,
			})
		}
	}

	cached, err := instance.ReadCachedEvents()
	if err != nil {
		return nil, nil, err
	}
	events = append(events, cached...)

	var earliestStart, latestEnd time.Time

	for _, event := range events {
		if earliestStart.IsZero() || earliestStart.After(event.Props.Start) {
			earliestStart = event.Props.Start
		}
		if latestEnd.IsZero() || latestEnd.Before(event.Props.End) {
			latestEnd = event.Props.End
		}
	}

	unsatisfiedRecurrences := []*Event{}

	recurrenceRange := timeRange
	if recurrenceRange.IsZero() {
		recurrenceRange.From = earliestStart
		recurrenceRange.To = latestEnd
	}
	for _, event := range events {
		if event.Props.Recurrence.IsThereRecurrence() {
			rruleSet, err := event.Props.GetRruleSet()
			if err != nil {
				log.Printf("warning: '%s' has an invalid recurrence set, and any recurrences were ignored: %s\n", event.Path, err)
				continue
			}
			recurrences := rruleSet.Between(recurrenceRange.From, recurrenceRange.To, true)
			if len(recurrences) > 0 {
				recurrences = recurrences[1:] // Ignore the first one because it already exists.
				for i, recurrence := range recurrences {
					newProps := event.Props
					newProps.Start = recurrence
					newProps.End = newProps.Start.Add(event.Props.End.Sub(event.Props.Start))

					p, err := NewEventPath(
						event.Path.Calendar(),
						fmt.Sprintf(".%s_%d", event.Path.Name(), i),
					)
					if err != nil {
						return nil, nil, err
					}

					events = append(events, Event{
						Path:     p,
						Props:    newProps,
						Type:     EventTypeRecurrence,
						Constant: true,
						Parent:   &event,
					})
				}
			}
			if timeRange.IsZero() && !rruleSet.After(recurrenceRange.To, false).IsZero() {
				unsatisfiedRecurrences = append(unsatisfiedRecurrences, &event)
			}
		}
	}

	// Don't delete the events until now, because in the case of a recurring event
	// that is outside the time range, but has children inside the time range.
	if !timeRange.IsZero() {
		events = FilterEvents(&events, func(e *Event) bool {
			return DoPeriodsMeet(e.Props.GetTimeRange(), timeRange)
		})
	}

	return events, unsatisfiedRecurrences, nil
}

// readDir reads a directory's events.
// The returned map's keys are the base filenames of the corresponding properties.
func (instance *Instance) readDir(dir string) (map[string]EventProperties, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	eventsProps := map[string]EventProperties{}

	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(dir, name)

		// Ignore dotfiles
		if strings.HasPrefix(name, ".") {
			continue
		}
		if entry.IsDir() {
			log.Printf("warning: ignoring calendar subdirectory '%s'. consider renaming it to start with a dot ('.'), to ignore it properly.\n", path)
			continue
		}

		props, err := parseEventFile(path)
		if err != nil {
			log.Printf("warning: event '%s' failed and was ignored: %s\n", path, err)
			continue
		}

		eventsProps[name] = props
	}

	return eventsProps, nil
}

func CreateInstance(root string) (*Instance, error) {
	config, err := ReadConfig(root)
	if err != nil {
		return nil, err
	}

	instance := &Instance{
		Root:   root,
		Config: config,
	}

	if err := instance.Work(); err != nil {
		return nil, err
	}

	return instance, nil
}
