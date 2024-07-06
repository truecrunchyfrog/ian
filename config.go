package ian

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

const CalendarConfigFilename string = ".config.toml"

type CalendarConfig struct {
	Sources map[string]CalendarSource
}

func getConfigPath(root string) string {
  return filepath.Join(root, CalendarConfigFilename)
}

func ReadConfig(root string) (CalendarConfig, error) {
	name := getConfigPath(root)

	buf, err := os.ReadFile(name)
	if err != nil {
		return CalendarConfig{}, err
	}

	var config CalendarConfig
	if _, err := toml.Decode(string(buf), &config); err != nil {
		return CalendarConfig{}, err
	}

  for name, source := range config.Sources {
    if source.Lifetime != "" {
      d, err := time.ParseDuration(source.Lifetime)
      if err != nil {
        return CalendarConfig{}, err
      }
      if d < 0 {
        return CalendarConfig{}, errors.New("in configuration source '" + name + "': lifetime cannot be negative.")
      }

      source._Lifetime = d
    }
  }


	return config, nil
}

func WriteConfig(root string, config CalendarConfig) error {
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(config); err != nil {
		return err
	}

	if err := os.WriteFile(getConfigPath(root), buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}
