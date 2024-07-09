package ian

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
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

const CooldownJournalFilename string = ".cooldown-journal.toml"

type SyncEventType int

const (
	SyncEventPing SyncEventType = 1 << iota
	SyncEventCreate
	SyncEventUpdate
	SyncEventDelete
)

type SyncEvent struct {
	Type    SyncEventType
	Files   string
	Message string
}

type SyncCooldownInfo struct {
	Cooldowns map[string]time.Time
}

// Sync is called whenever changes are made to event(s), and calls any configured commands.
func (instance *Instance) Sync(eventInfo SyncEvent) error {
  absRoot, err := filepath.Abs(instance.Root)
  if err != nil {
    return err
  }
	exportShellVars := fmt.Sprintf(
		`cd %s; export MESSAGE="%s"; export FILES="%s"; export TYPE=%d;`,
    absRoot,
		strings.ReplaceAll(eventInfo.Message, `"`, `\"`),
		strings.ReplaceAll(eventInfo.Files, `"`, `\"`),
		eventInfo.Type,
	)

	var cooldownJournal *SyncCooldownInfo
	var isJournalChanged bool

	for name, listener := range instance.Config.Sync.Listeners {
		if listener.Type == 0 || listener.Type&eventInfo.Type != 0 { // Type match
			ready := true

			if listener._Cooldown != 0 {
				if cooldownJournal == nil {
					buf, err := os.ReadFile(filepath.Join(instance.Root, CooldownJournalFilename))
					if err != nil && !os.IsNotExist(err) {
						return err
					}
					if _, err := toml.Decode(string(buf), &cooldownJournal); err != nil {
						return err
					}
				}

				// Don't need to check if map item exists with 'ok' because if it doesn't, lastChange will be 0 and it will work anyway.
				lastChange := cooldownJournal.Cooldowns[name]
				if now := time.Now(); lastChange.Add(listener._Cooldown).Before(now) {
					// Cooldown gone
					cooldownJournal.Cooldowns[name] = now
					isJournalChanged = true
				} else {
					// Still in cooldown
					ready = false
				}
			}

			if ready {
				cmd := exec.Command("sh", "-c", exportShellVars+listener.Command)
				if err := cmd.Run(); err != nil {
					buf := make([]byte, 32*1024)
					cmd.Stderr.Write(buf)
					log.Printf("warning: sync listener command '%s' exited unsuccessfully (%s). stderr: %s\n", name, err, string(buf))
				}
			}
		}
	}

	if isJournalChanged {
		buf := new(bytes.Buffer)
		if err := toml.NewEncoder(buf).Encode(cooldownJournal); err != nil {
			return err
		}
		os.WriteFile(filepath.Join(instance.Root, CooldownJournalFilename), buf.Bytes(), 0644)
	}

	return nil
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

	instance.ReadEvent(path)

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
		Path:   relPath,
		Props:  props,
		Cached: isPathInCache(relPath),
	})

	return nil
}

// ReadEvents reads and parses all events in the instance directory (recursively).
// Dotfiles are ignored.
func (instance *Instance) ReadEvents() error {
	instance.Events = []Event{}

  if err := instance.readDir(instance.Root); err != nil {
		return err
	}

	return nil
}

func (instance *Instance) readDir(dir string) error {
	entries, err := os.ReadDir(dir)
  if err != nil {
    return err
  }

	for _, entry := range entries {
    path := filepath.Join(dir, entry.Name())
    if entry.IsDir() {
      instance.readDir(path)
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

		instance.ReadEvent(relPath)
	}

	return nil
}

func (instance *Instance) GetEvent(relPath string) (*Event, error) {
	for _, ev := range instance.Events {
		if ev.Path == relPath {
			return &ev, nil
		}
	}
	return nil, fmt.Errorf("no such event '%s'", relPath)
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

	// Cached is true if the event should not be changed and is a cached event.
	Cached bool
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

func isUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
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
	case p.Url != "" && !isUrl(p.Url):
		return errors.New("URL is invalid")
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
