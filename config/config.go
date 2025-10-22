package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	MaxFiles             int `toml:"max-files"`
	MaxFileNameLength    int `toml:"max-file-name-length"`
	ImageBorderWidth     int `toml:"image-border-width"`
	ImageBorderHeight    int `toml:"image-border-height"`
	SelectorBorderWidth  int `toml:"selector-border-width"`
	SelectorBorderHeight int `toml:"selector-border-Height"`
}

var C *Config

func Load(path string) error {
	C = &Config{}
	_, err := toml.DecodeFile(path, C)
	if err != nil {
		return err
	}
	return nil
}
