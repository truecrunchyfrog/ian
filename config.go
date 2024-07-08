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
}

type ContainerConfig struct {
	Color color.RGBA
}

func getConfigPath(root string) string {
	return filepath.Join(root, CalendarConfigFilename)
}

func ReadConfig(root string) (Config, error) {
	name := getConfigPath(root)

	buf, err := os.ReadFile(name)
	if err != nil {
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

			source._Lifetime = d
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
