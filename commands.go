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
	Title     string
	Group     ISCPGroup
	ParamType ParamType
}

// QueryCommand generates the "xxxQSTN" command for this Command.
func (c *Command) QueryCommand() ISCPCommand {
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

func formatOnOffToggle(raw interface{}) (string, error) {
	result, err := formatToggle(raw)
	if err == nil {
		return result, err
	}
	return formatOnOff(raw)
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

// A CommandSet represnts a set of known/supported commands
// and can be used to convert the "friendly" version to ISCP and vice-versa.
type CommandSet interface {
	// LookupCommand finds the command definition for an ISCP command
	LookupCommand(ISCPCommand) (Command, error)
	// CreateCommand creates an ISCP command
	// for the given friendlyName name and parameter.
	// An error is returned if the name or parameter is invalid.
	CreateCommand(string, interface{}) (ISCPCommand, error)
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

func (b *basicCommandSet) LookupCommand(command ISCPCommand) (Command, error) {
	group, _ := SplitISCP(command)
	c, ok := b.byGroup[group]
	if !ok {
		return Command{}, errors.New("unknown ISCP command")
	}
	return c, nil
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
