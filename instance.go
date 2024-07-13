package ian

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
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
	return os.RemoveAll(filepath.Join(instance.Root, SanitizePath(name)))
}

// CreateEvent creates an event in the instance.
// containerDir is a directory relative to the root that the event will be placed in (leave empty to set it directly in the root).
func (instance *Instance) CreateEvent(props EventProperties, containerDir string) (*Event, error) {
	containerDir = SanitizePath(containerDir)

	path, err := instance.getAvailableFilepath(
		filepath.Join(containerDir, SanitizePath(props.FormatName())))
	if err != nil {
		return nil, err
	}

	event := Event{
		Path:  path,
		Props: props,
	}

	if err := event.Write(instance); err != nil {
		return nil, err
	}

	return &event, nil
}

func (instance *Instance) getAvailableFilepath(originalPath string) (string, error) {
	var pathSuffix string

	for i := 2; ; i++ {
		if i > 10 {
			return "", errors.New("cannot create file with that name: tried to add numerical suffix up to 10, but files by those names already exist.")
		}

		if _, err := os.Stat(filepath.Join(instance.Root, originalPath+pathSuffix)); err == nil {
			// File already exists
			pathSuffix = "_" + fmt.Sprint(i)
			continue
		}

		// Safe to write file
		return originalPath + pathSuffix, nil
	}
}

func (instance *Instance) readEvent(relPath string) (Event, error) {
	relPath = SanitizePath(relPath)
	path := filepath.Join(instance.Root, relPath)

	props, err := parseEventFile(path)
	if err != nil {
		return Event{}, err
	}

	if err := props.Validate(); err != nil {
		return Event{}, fmt.Errorf("failed validation: %s", err)
	}

	eType := EventTypeNormal

	if IsPathInCache(relPath) {
		eType = EventTypeCache
	}

	return Event{
		Path:     relPath,
		Props:    props,
		Type:     eType,
		Constant: eType == EventTypeCache,
	}, nil
}

// ReadEvents reads all events in the instance during the time range, and parses their recurrences.
// If the time range is empty (From.IsZero() && To.IsZero()), then all events are shown,
// and recurrences are shown within the range of the normal events.
func (instance *Instance) ReadEvents(timeRange TimeRange) ([]Event, []*Event, error) {
	events, err := instance.readDir(instance.Root)
	if err != nil {
		return nil, nil, err
	}

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
					events = append(events, Event{
						Path:     fmt.Sprintf("%s_%d", path.Join(path.Dir(event.Path), "."+path.Base(event.Path)), i),
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

// readDir reads a directory's events and directories recursively.
func (instance *Instance) readDir(dir string) ([]Event, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	events := []Event{}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		// Ignore dotfiles
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if entry.IsDir() {
			evs, err := instance.readDir(path)
			if err != nil {
				return nil, err
			}
			events = append(events, evs...)
			continue
		}
		relPath, err := filepath.Rel(instance.Root, path)
		if err != nil {
			log.Printf("warning: path for '%s' failed and was ignored: %s\n", path, err)
			continue
		}

		event, err := instance.readEvent(relPath)
		if err != nil {
			log.Printf("warning: event '%s' failed and was ignored: %s\n", path, err)
			continue
		}
		events = append(events, event)
	}

	return events, nil
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
