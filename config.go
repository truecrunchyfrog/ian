package ian

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/arran4/golang-ical"
)

const CalendarConfigFilename string = ".config.toml"

type CalendarConfig struct {
	Import []CalendarImport
}

type CalendarImport struct {
	Name    string
	Source  string
	NoCache bool
  // Type can be:
  // "native" for a native, dynamic ian calendar.
  // "ical" for a static HTTP iCalendar.
  // "caldav" for a dynamic CalDAV.
  Type string
}

func (i *CalendarImport) Import() ([]CalendarEvent, error) {
  switch i.Type {
  case "native": // TODO
		return nil, errors.New("native calendars not yet supported")
  case "caldav": // TODO
		return nil, errors.New("caldav calendars not yet supported")
  case "ical":
    resp, err := http.Get(i.Source)
    if err != nil {
      return nil, err
    }
    defer resp.Body.Close()
    ical, err := ics.ParseCalendar(resp.Body)
    if err != nil {
      return nil, err
    }
    cal, err := FromIcal(ical)
    if err != nil {
      return nil, err
    }
    return cal, nil
  default:
    return nil, errors.New("invalid calendar type '" + i.Type + "'")
  }
}

func ReadConfig(rootDir string) (CalendarConfig, error) {
	name := filepath.Join(rootDir, CalendarConfigFilename)

	buf, err := os.ReadFile(name)
	if err != nil {
		return CalendarConfig{}, err
	}

	var config CalendarConfig
	if _, err := toml.Decode(string(buf), &config); err != nil {
		return CalendarConfig{}, nil
	}

	return config, nil
}
