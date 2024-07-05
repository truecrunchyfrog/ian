package ian

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

const CalendarConfigFilename string = ".config.toml"

type CalendarConfig struct {
	Import []CalendarSource
}

func ReadConfig(root string) (CalendarConfig, error) {
	name := filepath.Join(root, CalendarConfigFilename)

	buf, err := os.ReadFile(name)
	if err != nil {
		return CalendarConfig{}, err
	}

	var config CalendarConfig
	if _, err := toml.Decode(string(buf), &config); err != nil {
		return CalendarConfig{}, nil
	}

  for _, source := range config.Import {
    if source.Lifetime != "" {
      d, err := time.ParseDuration(source.Lifetime)
      if err != nil {
        return CalendarConfig{}, err
      }
      if d < 0 {
        return CalendarConfig{}, errors.New("in configuration source '" + source.Name + "': lifetime cannot be negative.")
      }

      source._Lifetime = d
    }
  }

	return config, nil
}
