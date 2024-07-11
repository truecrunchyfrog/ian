package ian

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/teambition/rrule-go"
)

var Verbose bool

// SanitizePath escapes a string path. It prevents root traversal (/) and parent traversal (..), and just cleans it too.
func SanitizePath(path string) string {
	return filepath.Join("/", path)[1:]
}

type Instance struct {
	Root   string
	Config Config
}

// Work performs maintenance work and is run on every instance creation.
// It is used to e.g. update sources.
func (instance *Instance) Work() error {
	if err := instance.readEvents(); err != nil {
		return err
	}
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

	event.Write(instance)

	instance.readEvent(path)

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

func (instance *Instance) readEvent(relPath string, from, to time.Time) ([]Event, error) {
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
		Event{
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
		recurrences := rruleSet.Between(from, to, true)
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

func (instance *Instance) readEvents() ([]Event, error) {
	return instance.readDir(instance.Root)
}

func (instance *Instance) readDir(dir string) ([]Event, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			// Ignore dot directories and their contents, like '.git'.
			if !strings.HasPrefix(filepath.Base(dir), ".") {
				instance.readDir(path)
			}
			continue
		}
		// Ignore dotfiles
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		relPath, err := filepath.Rel(instance.Root, path)
		if err != nil {
			log.Println("warning: path for '"+path+"' failed and the event was ignored:", err)
			continue
		}

		instance.readEvent(relPath)
	}

	return nil
}

func (instance *Instance) GetEvent(relPath string) (*Event, error) {
	for _, ev := range instance.Events {
		if ev.Path == relPath {
			return ev, nil
		}
	}
	return nil, fmt.Errorf("no such event '%s'", relPath)
}

func (instance *Instance) FilterEvents(filter func(*Event) bool) []*Event {
	events := []*Event{}

	for _, event := range instance.Events {
		if filter(event) {
			events = append(events, event)
		}
	}

	return events
}

func CreateInstance(root string, performWork bool) (*Instance, error) {
	config, err := ReadConfig(root)
	if err != nil {
		return nil, err
	}

	instance := &Instance{
		Root:   root,
		Config: config,
		Events: []*Event{},
	}

	if performWork {
		if err := instance.Work(); err != nil {
			return nil, err
		}
	}

	return instance, nil
}

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

func CreateDir(name string) error {
	if err := os.MkdirAll(name, 0755); err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func CreateFileIfMissing(name string) error {
	CreateDir(filepath.Dir(name))
	if f, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err != nil {
		return err
	} else {
		f.Close()
		return nil
	}
}

// parseEventFile simply reads a file and parses it for properties.
func parseEventFile(path string) (EventProperties, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return EventProperties{}, err
	}

	var props EventProperties
	if _, err := toml.Decode(string(buf), &props); err != nil {
		return EventProperties{}, err
	}

	return props, nil
}
