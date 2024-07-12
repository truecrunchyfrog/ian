package ian

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/teambition/rrule-go"
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

func (instance *Instance) readEvent(relPath string, recurrenceTimeRange TimeRange) ([]Event, error) {
	relPath = SanitizePath(relPath)
	path := filepath.Join(instance.Root, relPath)

	props, err := parseEventFile(path)
	if err != nil {
		return nil, err
	}

	if err := props.Validate(); err != nil {
		return nil, fmt.Errorf("failed validation: %s", err)
	}

	events := []Event{
		{
			Path:     relPath,
			Props:    props,
			Constant: IsPathInCache(relPath),
		},
	}

	// RRule children
	if props.Rrule != "" {
		rruleSet, err := rrule.StrToRRuleSet(props.Rrule)
		if err != nil {
			log.Println("warning: '"+path+"' has an invalid RRule property and the event was ignored:", err)
			return nil, nil
		}
		recurrences := rruleSet.Between(recurrenceTimeRange.From, recurrenceTimeRange.To, true)
		if len(recurrences) > 0 {
			for i, recurrence := range recurrences[1:] {
				newProps := props
				newProps.Start = recurrence
				newProps.End = newProps.Start.Add(props.End.Sub(props.Start))
				events = append(events, Event{
					Path:     fmt.Sprintf(".fork/%s_%d", relPath, i),
					Props:    newProps,
					Constant: true,
					Parent:   &events[0],
				})
			}
		}
	}

	return events, nil
}

// ReadEvents reads all events in the instance. Recurring events (via rrule) will be limited to the bounds of recurrenceTimeRange.
// Normal events will not be filtered by recurrenceTimeRange.
func (instance *Instance) ReadEvents(recurrenceTimeRange TimeRange) ([]Event, error) {
	return instance.readDir(instance.Root, recurrenceTimeRange)
}

func (instance *Instance) readDir(dir string, recurrenceTimeRange TimeRange) ([]Event, error) {
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
      evs, err := instance.readDir(path, recurrenceTimeRange)
      if err != nil {
        return nil, err
      }
      events = append(events, evs...)
			continue
		}
		relPath, err := filepath.Rel(instance.Root, path)
		if err != nil {
			log.Printf("warning: path for '%s' failed and the event was ignored: %s\n", path, err)
			continue
		}

		evs, err := instance.readEvent(relPath, recurrenceTimeRange)
		if err != nil {
			log.Printf("warning: event '%s' failed and the event was ignored: %s\n", path, err)
			continue
		}
		events = append(events, evs...)
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
