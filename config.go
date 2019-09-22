package onkyoctl

import (
	"github.com/go-ini/ini"
)

const defaultPort = 60128

// Config hold configuration settings.
type Config struct {
	Host           string
	Port           int
	ConnectTimeout int
}

// DefaultConfig returns a Config struct with default values.
func DefaultConfig() *Config {
	return &Config{
		Port:           defaultPort,
		ConnectTimeout: 10,
	}
}

// ReadConfig reads configuration froin ini format from the given source.
// Source can be a path, an opened file or a []byte array.
func ReadConfig(source interface{}) (*Config, error) {
	iniValues, err := ini.Load(source)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	err = iniValues.MapTo(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
