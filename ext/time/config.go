package time

import (
	"github.com/bmuller/arrow"

	"github.com/paradime-io/gonja/config"
)

type Config struct {
	// Default format
	DateTimeFormat string
	// If defined, now returns this parsed value
	Now *arrow.Arrow
}

func NewConfig() *Config {
	return &Config{
		DateTimeFormat: "%Y-%m-%d",
		Now:            nil,
	}
}

func (cfg *Config) Inherit() config.Inheritable {
	return &Config{
		DateTimeFormat: cfg.DateTimeFormat,
		Now:            cfg.Now,
	}
}

// DefaultConfig is a configuration with default values
var DefaultConfig = NewConfig()
