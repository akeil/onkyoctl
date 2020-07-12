package onkyoctl

import (
	"fmt"
	"io/ioutil"

	"github.com/go-ini/ini"
	"gopkg.in/yaml.v2"
)

const defaultPort = 60128

// Config holds configuration settings.
type Config struct {
	Host             string
	Port             int
	AutoConnect      bool
	AllowReconnect   bool
	ReconnectSeconds int
	CommandFile      string
	Commands         CommandSet
	Log              Logger
}

// DefaultConfig returns a Config struct with default values.
func DefaultConfig() *Config {
	return &Config{
		Port:             defaultPort,
		AutoConnect:      false,
		AllowReconnect:   false,
		ReconnectSeconds: 5,
	}
}

// ReadConfig reads configuration from ini format from the given source.
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

	if cfg.CommandFile != "" {
		cmd, err := ReadCommands(cfg.CommandFile)
		if err != nil {
			return nil, err
		}
		cfg.Commands = cmd
	}

	return cfg, nil
}

// ReadCommands loads a CommandSet from a YAML file specified by the given
// path.
func ReadCommands(path string) (CommandSet, error) {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read commands: %v", err)
	}

	c := make([]Command, 0)
	err = yaml.Unmarshal(d, &c)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal commands YAML: %v", err)
	}

	return NewBasicCommandSet(c), nil
}
