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
)

var Verbose bool

// SanitizePath escapes a string path. It prevents root traversal (/) and parent traversal (..), and just cleans it too.
func SanitizePath(path string) string {
	return filepath.Join("/", path)[1:]
}

type Instance struct {
	Root   string
	Config Config
	Events []Event
}

// Work performs maintenance work and is run on every instance creation.
// It is used to e.g. update sources.
func (instance *Instance) Work() error {
	if err := instance.ReadEvents(); err != nil {
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

// CreateEvent creates an event in the instance and returns it.
// containerDir is a directory relative to the root that the event will be placed in (leave empty to set it directly in the root).
func (instance *Instance) CreateEvent(props EventProperties, containerDir string) error {
	containerDir = SanitizePath(containerDir)

	path, err := instance.getAvailableFilepath(
		filepath.Join(containerDir, SanitizePath(props.FormatName())))
	if err != nil {
		return err
	}

	(&Event{
		Path:  path,
		Props: props,
	}).Write(instance)

	instance.ReadEvent(path)

	return nil
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

func (instance *Instance) ReadEvent(relPath string) error {
	relPath = SanitizePath(relPath)
	path := filepath.Join(instance.Root, relPath)

	props, err := parseEventFile(path)
	if err != nil {
		log.Println("warning: parsing '"+path+"' failed and the event was ignored:", err)
		return nil
	}

  if err := props.Verify(); err != nil {
		log.Println("warning: '"+path+"' failed verification and the event was ignored:", err)
		return nil
  }

	// Delete old version if it exists:
	i := slices.IndexFunc(instance.Events, func(event Event) bool {
		return event.Path == relPath
	})
	if i != -1 {
		instance.Events = slices.Delete(instance.Events, i, i+1)
	}

	instance.Events = append(instance.Events, Event{
		Path:     relPath,
		Props:    props,
		Constant: isPathInCache(relPath),
	})

	return nil
}

// ReadEvents reads and parses all events in the instance directory (recursively).
// Dotfiles are ignored.
func (instance *Instance) ReadEvents() error {
	instance.Events = []Event{}

	err := filepath.WalkDir(instance.Root, func(path string, d os.DirEntry, err error) error {
		// Ignore directories and dotfiles
		if d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		relPath, err := filepath.Rel(instance.Root, path)
		if err != nil {
			log.Println("warning: path for '"+path+"' failed and the event was ignored:", err)
			return nil
		}

		instance.ReadEvent(relPath)

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (instance *Instance) FilterEvents(filter func(Event) bool) []Event {
	events := []Event{}

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
		Events: []Event{},
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

	// Constant is true if the event should not be changed.
	// This is used for static imported calendars.
	Constant bool
}

func (event *Event) GetRealPath(instance *Instance) string {
	return filepath.Join(instance.Root, event.Path)
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
	return fmt.Sprintf("%30s @ %s → %s (%s)",
		event.Props.Summary,
		event.Props.Start.Format(DefaultTimeLayout),
		event.Props.End.Format(DefaultTimeLayout),
		DurationToString(event.Props.End.Sub(event.Props.Start)),
	)
}

type EventProperties struct {
	Summary     string
	Description string
	Location    string
	Url         string

	Start  time.Time
	End    time.Time
	AllDay bool

	Created  time.Time
	Modified time.Time
}

func (p *EventProperties) Verify() error {
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
		fallthrough
	default:
		return nil
	}
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
