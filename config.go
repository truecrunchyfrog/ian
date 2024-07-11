package ian

import (
	"bytes"
	"errors"
	"image/color"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

const CalendarConfigFilename string = ".config.toml"

type Config struct {
	Calendars map[string]ContainerConfig
	Sources   map[string]CalendarSource
	Sync      SyncConfig
}

type ContainerConfig struct {
	Color color.RGBA
}

type SyncConfig struct {
	// Listeners contains a list of shell commands that are executed when certain calendar event changes occur.
	// Useful for synchronizing VCS, backups, or pinging an ian server.
	Listeners map[string]struct {
		// Command is run as a shell command when an event is updated, in the instance directory.
		//
		// Use $MESSAGE in the command to embed the message describing the event change.
		//
		// Use $FILES for a space-separated string with the affected file(s).
		//
		// Use $TYPE for the event type ID.
		//
		// Any stderr output from the command is printed to the user in the form of a warning.
		//
		// Example: 'git add . && git commit -m "$MESSAGE" && (git pull; git push)'
		Command string
		// Type is a bitmask that represents the event(s) to listen to.
		Type SyncEventType
		// Cooldown is parsed as a time.Duration, and is the duration that has to pass before the command is executed again, to prevent fast-paced command execution.
		Cooldown  string
		Cooldown_ time.Duration
	}
}

func getConfigPath(root string) string {
	return filepath.Join(root, CalendarConfigFilename)
}

func ReadConfig(root string) (Config, error) {
	name := getConfigPath(root)

	buf, err := os.ReadFile(name)
	if err != nil && !os.IsNotExist(err) {
		return Config{}, err
	}

	var config Config
	if _, err := toml.Decode(string(buf), &config); err != nil {
		return Config{}, err
	}

	for name, source := range config.Sources {
		if source.Lifetime != "" {
			d, err := time.ParseDuration(source.Lifetime)
			if err != nil {
				return Config{}, err
			}
			if d < 0 {
				return Config{}, errors.New("in configuration source '" + name + "': lifetime cannot be negative.")
			}

			source.Lifetime_ = d
			config.Sources[name] = source
		}
	}

	for name, listener := range config.Sync.Listeners {
		if listener.Cooldown != "" {
			d, err := time.ParseDuration(listener.Cooldown)
			if err != nil {
				return Config{}, err
			}
			if d < 0 {
				return Config{}, errors.New("in configuration listener '" + name + "': cooldown cannot be negative.")
			}

			listener.Cooldown_ = d
			config.Sync.Listeners[name] = listener
		}
	}

	return config, nil
}

func WriteConfig(root string, config Config) error {
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(config); err != nil {
		return err
	}

	if err := os.WriteFile(getConfigPath(root), buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func (conf *Config) GetContainerConfig(container string) (*ContainerConfig, error) {
	for name, cal := range conf.Calendars {
		if name == container {
			return &cal, nil
		}
	}
	return nil, errors.New("calendar config for '" + container + "' does not exist")
}

func (conf *ContainerConfig) GetColor() color.RGBA {
	if r, g, b, _ := conf.Color.RGBA(); r+g+b == 0 {
		return color.RGBA{255, 255, 255, 255}
	}
	return conf.Color
}
