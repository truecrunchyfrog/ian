package ian

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	ics "github.com/arran4/golang-ical"
)

const CacheDirName string = "cache"
const CacheJournalFileName string = ".cache-journal.toml"
const DefaultCacheLifetime time.Duration = 2 * time.Hour

type CalendarSource struct {
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
	Sources map[string]CacheJournalSource
}

type CacheJournalSource struct {
	LastUpdate time.Time
}

func (i *CalendarSource) Import(name string) ([]EventProperties, error) {
	if Verbose {
		log.Printf("source '%s' is being imported as '%s'\n", name, i.Type)
	}
	switch i.Type {
	case "native": // TODO
		return nil, errors.New("native calendars not yet supported")
	case "caldav": // TODO
		return nil, errors.New("caldav calendars not yet supported")
	case "ical":
		if Verbose {
			log.Printf("downloading iCalendar '%s'\n", i.Source)
		}
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

func (i *CalendarSource) ImportAndUse(instance *Instance, name string) error {
	cal, err := i.Import(name)
	if err != nil {
		return err
	}

	instance.CacheEvents(name, cal)

	return nil
}

func (instance *Instance) DeleteCache() error {
	return os.RemoveAll(instance.getCacheDir())
}

func (instance *Instance) CleanSources() error {
	if err := instance.DeleteCache(); err != nil {
		return err
	}

	if err := instance.UpdateSources(); err != nil {
		return err
	}
	return nil
}

// UpdateSources updates the configured sources according to their lifetimes.
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
	var isJournalChanged bool
	if journal.Sources == nil {
		journal.Sources = map[string]CacheJournalSource{}
		isJournalChanged = true
	}

	now := time.Now()
	unsatisfiedSources := maps.Clone(instance.Config.Sources)

	for name, journalSource := range journal.Sources {
		source, ok := unsatisfiedSources[name]
		if !ok {
			log.Printf("warning: in cache journal '%s': source with the name '%s' does not exist.\n", path, name)
			continue
		}

		var lifetime time.Duration
		if source.Lifetime != "" {
			lifetime = source._Lifetime
		} else {
			lifetime = DefaultCacheLifetime
		}

		if journalSource.LastUpdate.Add(lifetime).Before(now) {
			// Lifetime expired, update the source.
      journalSource.LastUpdate = now
      journal.Sources[name] = journalSource
			isJournalChanged = true
			if err := source.ImportAndUse(instance, name); err != nil {
				return err
			}
		}

		delete(unsatisfiedSources, name)
	}

	// Unsatisfied source (sources missing from journal) are updated and added to the journal.

	if len(unsatisfiedSources) != 0 {
		log.Printf("adding %d source(s) to journal\n", len(unsatisfiedSources))
	}

	for name, source := range unsatisfiedSources {
		if Verbose {
			log.Printf("source '%s' is not provided in journal. it will be updated and added.\n", name)
		}

		if err := source.ImportAndUse(instance, name); err != nil {
			return err
		}

		journal.Sources[name] = CacheJournalSource{
			LastUpdate: now,
		}
		isJournalChanged = true
	}

	// Write updated journal

	if isJournalChanged {
    if Verbose {
      log.Printf("updating journal:\n%s\n", journal)
    }

		bufOut := new(bytes.Buffer)
    bufOut.WriteString(
      fmt.Sprintf(
        "# This file is automatically generated and managed.\n# Last change: %s\n\n",
        now.Format(DefaultTimeLayout),
      ),
    )
		if err := toml.NewEncoder(bufOut).Encode(journal); err != nil {
			return err
		}

		if err := os.WriteFile(path, bufOut.Bytes(), 0644); err != nil {
			return err
		}
	}

	return nil
}

func (instance *Instance) getCacheDir() string {
	return filepath.Join(instance.Root, CacheDirName)
}

func (instance *Instance) CacheEvent(name string, props EventProperties) error {
	return instance.CreateEvent(props, filepath.Join(CacheDirName, SanitizePath(name)))
}

// CacheEvents collectively caches a list of events under a certain directory.
func (instance *Instance) CacheEvents(name string, eventsProps []EventProperties) error {
	// First empty the specified cache directory
	instance.clearDir(filepath.Join(CacheDirName, name))

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
