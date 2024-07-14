package ian

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

var Verbose bool

var TimeZone *time.Location

func GetTimeZone() *time.Location {
	if TimeZone != nil {
		return TimeZone
	}

	timeZoneFlag := viper.GetString("timezone")

	if timeZoneFlag == "" {
		return time.Local
	}

	t1, err := time.Parse("MST", timeZoneFlag)

	if err == nil {
		TimeZone = t1.Location()
	} else {
		t2, err := time.Parse("-0700", timeZoneFlag)
		if err != nil {
			log.Fatal("invalid time zone '" + timeZoneFlag + "'")
		}

		TimeZone = t2.Location()
	}

	return TimeZone
}

// SanitizeFilepath escapes a filepath. It prevents root traversal (/) and parent traversal (..), and just cleans it too.
func SanitizeFilepath(p string) string {
	return strings.TrimPrefix(filepath.Join(string(filepath.Separator), p), string(filepath.Separator))
}

// SanitizePath escapes a path. It prevents root traversal (/) and parent traversal (..), and just cleans it too.
func SanitizePath(p string) string {
	return strings.TrimPrefix(path.Join("/", p), "/")
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

func GenerateUid() string {
	return strings.ToUpper(uuid.New().String())
}
