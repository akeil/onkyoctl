package onkyoctl

import (
	"errors"
	"strconv"
	"strings"
)

// ISCPGroup is the 3-digit ISCP command group, e.g. "PWR" or "MVL".
type ISCPGroup string

// An ISCPCommand is a low-level command like PWR01 (power on)
// or MVLUP (master volume up).
type ISCPCommand string

// ParamType is the kind of parameter expcted by a Command.
type ParamType string

const (
	paramOnOff       ParamType = "onOff"
	paramOnOffToggle ParamType = "onOffToggle"
	// lookup
	// lookupToggle

	queryParam = "QSTN"
)

// SplitISCP splits an ISCP command into group and parameter.
func SplitISCP(command ISCPCommand) (ISCPGroup, string) {
	s := string(command)
	group := ISCPGroup(s[0:3])
	param := s[3:]

	return group, param
}

// Command is the "friendly" wrapper around an ISCP command group.
type Command struct {
	Name      string
	Group     ISCPGroup
	ParamType ParamType
}

// CreateQuery generates the "xxxQSTN" command for this Command.
func (c *Command) CreateQuery() ISCPCommand {
	return ISCPCommand(string(c.Group) + queryParam)
}

// CreateCommand creates an ISCP command with the given parameter.
// An error is returned if the parameter is invalid.
func (c *Command) CreateCommand(param interface{}) (ISCPCommand, error) {
	p, err := c.formatParam(param)
	if err != nil {
		return "", err
	}
	return ISCPCommand(string(c.Group) + p), nil
}

func (c *Command) formatParam(raw interface{}) (string, error) {
	switch c.ParamType {
	case paramOnOff:
		return formatOnOff(raw)
	case paramOnOffToggle:
		return formatOnOffToggle(raw)
	}
	return "", errors.New("invalid param type")
}

// ParseParam converts the ISCP param value to the friendly version.
func (c *Command) ParseParam(raw string) (string, error) {
	switch c.ParamType {
	case paramOnOff:
		return parseOnOff(raw)
	case paramOnOffToggle:
		return parseOnOffToggle(raw)
	}
	return "", errors.New("invalid param type")
}

func formatOnOff(raw interface{}) (string, error) {
	var result string

	switch val := raw.(type) {
	case bool:
		if val {
			result = "01"
		} else {
			result = "00"
		}
	case int,
		int8,
		int16,
		int32,
		int64,
		uint,
		uint8,
		uint16,
		uint32,
		uint64,
		float32,
		float64:
		if val == 1 || val == 1.0 {
			result = "01"
		} else if val == 0 || val == 0.0 {
			result = "00"
		}
	case string:
		find := strings.ToLower(val)
		if find == "on" {
			result = "01"
		} else if find == "off" {
			result = "00"
		} else {
			b, convErr := strconv.ParseBool(val)
			if convErr == nil {
				return formatOnOff(b)
			}
			i, convErr := strconv.Atoi(val)
			if convErr == nil {
				return formatOnOff(i)
			}
		}
	}

	if result == "" {
		return "", errors.New("invalid parameter")
	}
	return result, nil
}

func parseOnOff(raw string) (string, error) {
	switch raw {
	case "00":
		return "off", nil
	case "01":
		return "on", nil
	default:
		return "", errors.New("invalid parameter")
	}
}

func formatOnOffToggle(raw interface{}) (string, error) {
	result, err := formatToggle(raw)
	if err == nil {
		return result, err
	}
	return formatOnOff(raw)
}

func parseOnOffToggle(raw string) (string, error) {
	parsed, err := parseToggle(raw)
	if err == nil {
		return parsed, err
	}
	return parseOnOff(raw)
}

func formatToggle(raw interface{}) (string, error) {
	s, ok := raw.(string)
	if ok {
		s = strings.ToLower(s)
		if s == "" || s == "toggle" || s == "tg" {
			return "TG", nil
		}
	}
	return "", errors.New("invalid parameter")
}

func parseToggle(raw string) (string, error) {
	if raw == "TG" {
		return "toggle", nil
	}
	return "", errors.New("invalid parameter")
}

// A CommandSet represnts a set of known/supported commands
// and can be used to convert the "friendly" version to ISCP and vice-versa.
type CommandSet interface {
	// ReadCommand finds the command definition for an ISCP command
	// and converts the parameter.
	ReadCommand(ISCPCommand) (string, string, error)
	// CreateCommand creates an ISCP command
	// for the given friendlyName name and parameter.
	// An error is returned if the name or parameter is invalid.
	CreateCommand(string, interface{}) (ISCPCommand, error)
	// CreateQuery creates a QSTN command for the given friendly name.
	CreateQuery(string) (ISCPCommand, error)
}

type basicCommandSet struct {
	byGroup map[ISCPGroup]Command
	byName  map[string]Command
}

// NewBasicCommandSet creates a new CommandSet
// from the given list of command definitions.
func NewBasicCommandSet(commands []Command) CommandSet {
	byGroup := make(map[ISCPGroup]Command)
	byName := make(map[string]Command)
	for _, c := range commands {
		if c.Group != "" {
			byGroup[c.Group] = c
		}
		if c.Name != "" {
			byName[c.Name] = c
		}
	}

	return &basicCommandSet{
		byGroup: byGroup,
		byName:  byName,
	}
}

func (b *basicCommandSet) ReadCommand(command ISCPCommand) (string, string, error) {
	group, param := SplitISCP(command)
	c, ok := b.byGroup[group]
	if !ok {
		return "", "", errors.New("unknown ISCP command")
	}

	value, err := c.ParseParam(param)
	if err != nil {
		return "", "", err
	}
	return c.Name, value, nil
}

func (b *basicCommandSet) ForName(name string) (Command, error) {
	c, ok := b.byName[name]
	if !ok {
		return Command{}, errors.New("unknown command")
	}
	return c, nil
}

func (b *basicCommandSet) CreateCommand(name string, param interface{}) (ISCPCommand, error) {
	c, err := b.ForName(name)
	if err != nil {
		return "", err
	}
	return c.CreateCommand(param)
}

func (b *basicCommandSet) CreateQuery(name string) (ISCPCommand, error) {
	c, err := b.ForName(name)
	if err != nil {
		return "", err
	}
	return c.CreateQuery(), nil
}
