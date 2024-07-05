package ian

import (
	"errors"
	"net/http"
	"path/filepath"

	ics "github.com/arran4/golang-ical"
)

const CacheDirName string = "cache"

type CalendarSource struct {
	Name    string
	Source  string
  // Type can be:
  // "native" for a native, dynamic ian calendar.
  // "caldav" for a dynamic CalDAV.
  // "ical" for a static HTTP iCalendar.
  Type string
}

func (i *CalendarSource) Import() ([]EventProperties, error) {
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

func (instance *Instance) getCacheDir() string {
	return filepath.Join(instance.Root, CacheDirName)
}

func (instance *Instance) CacheEvent(name string, props EventProperties) error {
  return instance.CreateEvent(props, filepath.Join(instance.getCacheDir(), SanitizePath(name)))
}

// CacheEvents collectively caches a list of events under a certain directory.
func (instance *Instance) CacheEvents(name string, eventsProps []EventProperties) error {
  for _, props := range eventsProps {
    if err := instance.CacheEvent(name, props); err != nil {
      return err
    }
  }

	return nil
}

func isPathInCache(relPath string) bool {
  parts := filepath.SplitList(relPath)
  return len(parts) >= 2 && parts[0] == CacheDirName
}
