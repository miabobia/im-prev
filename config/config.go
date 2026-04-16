package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	MaxFiles                   int     `toml:"max-files"`
	MaxFileNameLength          int     `toml:"max-file-name-length"`
	ImageBorderWidth           int     `toml:"image-border-width"`
	ImageBorderHeight          int     `toml:"image-border-height"`
	SelectorBorderWidth        int     `toml:"selector-border-width"`
	SelectorBorderHeight       int     `toml:"selector-border-Height"`
	ChafaMaxWidth              int     `toml:"chafa-max-width"`
	ChafaMaxHeight             int     `toml:"chafa-max-height"`
	ChafaDefaultSymbols        string  `toml:"chafa-default-symbols"`
	FileViewBufferSize         int     `toml:"file-view-buffer-size"`
	ImageWidthRatio            float32 `toml:"image-width-ratio"`
	ImageHeightRatio           float32 `toml:"image-height-ratio"`
	SelectorWidthRatio         float32 `toml:"selector-width-ratio"`
	SelectorHeightRatio        float32 `toml:"selector-height-ratio"`
	ImageFullscreenWidthRatio  float32 `toml:"fullscreen-image-width-ratio"`
	ImageFullscreenHeightRatio float32 `toml:"fullscreen-image-height-ratio"`
	MinimumScreenWidth         int     `toml:"minimum-screen-width"`
	MinimumScreenHeight        int     `toml:"minimum-screen-height"`
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
