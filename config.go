package ian

import (
	"os"
	"path/filepath"

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

	return config, nil
}
