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

	t, err := time.LoadLocation(timeZoneFlag)
	if err != nil {
		log.Fatalf("invalid time zone '%s': %s", timeZoneFlag, err)
	}

	TimeZone = t

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
	if err := CreateDir(filepath.Dir(name)); err != nil {
		return err
	}
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

	props.Start = props.Start.Truncate(time.Second)
	props.End = props.End.Truncate(time.Second)

	props.Created = props.Created.Truncate(time.Second)
	props.Modified = props.Modified.Truncate(time.Second)

	return props, nil
}

func GenerateUid() string {
	return strings.ToUpper(uuid.New().String())
}
