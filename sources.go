package ian

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/BurntSushi/toml"
	ics "github.com/arran4/golang-ical"
)

const CacheDirName string = "cache"
const CacheJournalFileName string = ".cache-journal.toml"
const DefaultCacheLifetime time.Duration = 2 * time.Hour

type CalendarSource struct {
	Name   string
	Source string
	// Type can be:
	// "native" for a native, dynamic ian calendar.
	// "caldav" for a dynamic CalDAV.
	// "ical" for a static HTTP iCalendar.
	Type string
	// Parsed with time.ParseDuration...
	Lifetime string
	// and inserted here:
	_Lifetime time.Duration
}

type CacheJournal struct {
	Sources []struct {
		Name       string
		LastUpdate time.Time
	}
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

func (i *CalendarSource) ImportAndUse(instance *Instance) error {
	cal, err := i.Import()
	if err != nil {
		return err
	}

	instance.CacheEvents(i.Name, cal)

	return nil
}

func (instance *Instance) UpdateSources() error {
	path := filepath.Join(instance.getCacheDir(), CacheJournalFileName)
  CreateFileIfMissing(path)
	buf, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var journal CacheJournal
	if _, err := toml.Decode(string(buf), &journal); err != nil {
		return err
	}

	now := time.Now()
	unsatisfiedSources := slices.Clone(instance.Config.Import)

	for _, journalSource := range journal.Sources {
		i := slices.IndexFunc(unsatisfiedSources, func(src CalendarSource) bool {
			return src.Name == journalSource.Name
		})

		if i == -1 {
			log.Printf("warning: in cache journal '%s': journal source with the name '%s' does not exist or used twice", path, journalSource.Name)
			continue
		}

		source := unsatisfiedSources[i]

		var lifetime time.Duration
		if source.Lifetime != "" {
			lifetime = source._Lifetime
		} else {
			lifetime = DefaultCacheLifetime
		}

		if journalSource.LastUpdate.Add(lifetime).Before(now) {
			// Lifetime expired, update the source.
			source.ImportAndUse(instance)
		}

		unsatisfiedSources = slices.Delete(unsatisfiedSources, i, i+1)
	}

	// Unsatisfied source (sources missing from journal) are updated and added to the journal.

	for _, unsatisfied := range unsatisfiedSources {
    unsatisfied.ImportAndUse(instance)

		journal.Sources = append(journal.Sources, struct {
			Name       string
			LastUpdate time.Time
		}{
			Name:       unsatisfied.Name,
			LastUpdate: now,
		})
	}

	// Write updated journal

	bufOut := new(bytes.Buffer)
	if err := toml.NewEncoder(bufOut).Encode(journal); err != nil {
		return err
	}

	if err := os.WriteFile(path, bufOut.Bytes(), 0644); err != nil {
		return err
	}

	return nil
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
