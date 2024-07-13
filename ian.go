package ian

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
)

var Verbose bool

// SanitizePath escapes a string path. It prevents root traversal (/) and parent traversal (..), and just cleans it too.
func SanitizePath(path string) string {
	return filepath.Join("/", path)[1:]
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
