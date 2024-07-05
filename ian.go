package ian

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type CalendarEvent struct {
	Summary     string
	Description string
	Location    string
	Url         string

	Start time.Time
	End   time.Time

	Created  time.Time
	Modified time.Time

	// Constant is true if the event should not be changed.
	// This is used for static imported calendars.
	Constant bool
}

func (event *CalendarEvent) String() string {
	return fmt.Sprintf("%30s @ %s → %s (%s)",
		event.Summary,
		event.Start.Format(DefaultTimeLayout),
		event.End.Format(DefaultTimeLayout),
		event.End.Sub(event.Start),
	)
}

// GetFilename decides the filename for an event based on the event's name.
// "My event" -> "my-event"
func (event *CalendarEvent) GetFilename() string {
	name := event.Summary
	// Lowercase
	name = strings.ToLower(name)
	// Dashes instead of spaces
	name = strings.ReplaceAll(name, " ", "-")
	// Underscores instead of dots
	name = strings.ReplaceAll(name, ".", "_")
	return name
}

func CreateEventFile(event CalendarEvent, dir string) error {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(event)
	if err != nil {
		return err
	}

	filepath := filepath.Join(dir, event.GetFilename())
	var filepathSuffix string

	for i := 2; ; i++ {
		if i > 10 {
			return errors.New("cannot create file: tried to add numerical suffix up to 10, but files by those names already exist.")
		}

		if _, err := os.Stat(filepath + filepathSuffix); err == nil {
			// File already exists
			filepathSuffix = "_" + fmt.Sprint(i)
			continue
		}

		// Safe to write file
		filepath += filepathSuffix
		break
	}

	if err := os.WriteFile(filepath, buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func ReadEventFile(filepath string) (CalendarEvent, error) {
	buf, err := os.ReadFile(filepath)
	if err != nil {
		return CalendarEvent{}, err
	}

	var event CalendarEvent
	if _, err := toml.Decode(string(buf), &event); err != nil {
		return CalendarEvent{}, err
	}

	return event, nil
}

// ReadEventDir reads and parses the events of a calendar directory non-recursively (only the provided directory).
func ReadEventDir(dir string) ([]CalendarEvent, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var events []CalendarEvent

	for _, entry := range entries {
		// Ignore directories and dotfiles
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		name := filepath.Join(dir, entry.Name())

		event, err := ReadEventFile(name)
		if err != nil {
			log.Println("warning: '"+name+"' was ignored:", err)
			continue
		}

		events = append(events, event)
	}

	return events, nil
}
