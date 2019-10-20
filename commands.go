package onkyoctl

import (
	"fmt"
	"math"
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
	// OnOff commands accept only on/off as parameter.
	OnOff        ParamType = "onOff"
	// OnOffToggle commands work like on/off but accept an additional "toggle".
	OnOffToggle  ParamType = "onOffToggle"
	// Enum based commands expect parameters from a list of values.
	Enum         ParamType = "enum"
	// EnumToggle works like Enum but accepts an additional toggle/cycle parameter.
	EnumToggle   ParamType = "enumToggle"
	// IntRange accepts an integer value with min and max values.
	IntRange     ParamType = "intRange"
	// IntRangeEnum accepts integers and additional values from a list.
	IntRangeEnum ParamType = "intRangeEnum"

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
	Lookup    map[string]string
	Lower     int
	Upper     int
	Scale     int
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
	case OnOff:
		return formatOnOff(raw)
	case OnOffToggle:
		return formatOnOffToggle(raw)
	case Enum:
		return formatEnum(c.Lookup, raw)
	case EnumToggle:
		return formatEnumToggle(c.Lookup, raw)
	case IntRange:
		return formatIntRange(c.Lower, c.Upper, c.Scale, raw)
	case IntRangeEnum:
		return formatIntRangeEnum(c.Lower, c.Upper, c.Scale, c.Lookup, raw)
	}

	return "", fmt.Errorf("unsupported param type %q", c.ParamType)
}

// ParseParam converts the ISCP param value to the friendly version.
func (c *Command) ParseParam(raw string) (string, error) {
	switch c.ParamType {
	case OnOff:
		return parseOnOff(raw)
	case OnOffToggle:
		return parseOnOffToggle(raw)
	case Enum:
		return parseEnum(c.Lookup, raw)
	case EnumToggle:
		return parseEnumToggle(c.Lookup, raw)
	case IntRange:
		return parseIntRange(c.Lower, c.Upper, c.Scale, raw)
	case IntRangeEnum:
		return parseIntRangeEnum(c.Lower, c.Upper, c.Scale, c.Lookup, raw)
	}
	return "", fmt.Errorf("unsupported param type %q", c.ParamType)
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
		return "", fmt.Errorf("invalid parameter %q", raw)
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
		return "", fmt.Errorf("invalid parameter %q", raw)
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

func formatEnum(lookup map[string]string, raw interface{}) (string, error) {
	s := fmt.Sprintf("%v", raw)
	s = strings.ToLower(s)
	for key, value := range lookup {
		if value == s {
			return key, nil
		}
	}
	return "", fmt.Errorf("invalid parameter %q", raw)
}

func parseEnum(lookup map[string]string, raw string) (string, error) {
	value, ok := lookup[raw]
	if ok {
		return value, nil
	}
	return "", fmt.Errorf("invalid parameter %q", raw)
}

func formatEnumToggle(lookup map[string]string, raw interface{}) (string, error) {
	parsed, err := formatToggle(raw)
	if err == nil {
		return parsed, err
	}
	return formatEnum(lookup, raw)
}

func parseEnumToggle(lookup map[string]string, raw string) (string, error) {
	value, err := parseToggle(raw)
	if err == nil {
		return value, err
	}
	return parseEnum(lookup, raw)
}

func formatIntRange(lower, upper, scale int, raw interface{}) (string, error) {
	// conversion
	var numeric float64
	switch val := raw.(type) {
	case int:
		numeric = float64(val)
	case int8:
		numeric = float64(val)
	case int16:
		numeric = float64(val)
	case int32:
		numeric = float64(val)
	case int64:
		numeric = float64(val)
	case uint:
		numeric = float64(val)
	case uint8:
		numeric = float64(val)
	case uint16:
		numeric = float64(val)
	case uint32:
		numeric = float64(val)
	case uint64:
		numeric = float64(val)
	case float32:
		numeric = float64(val)
	case float64:
		numeric = val

	case string:
		var convErr error
		numeric, convErr = strconv.ParseFloat(val, 64)
		if convErr != nil {
			return "", convErr
		}
	default:
		return "", fmt.Errorf("invalid parameter %q", raw)
	}

	// bounds check
	if numeric < float64(lower) || numeric > float64(upper) {
		return "", fmt.Errorf("invalid parameter %q", raw)
	}

	if scale == 0 {
		scale = 1
	}
	scaled := numeric * float64(scale)
	rounded := math.Round(scaled)
	// rounding should not change the value
	if rounded != scaled {
		return "", fmt.Errorf("invalid parameter %q", raw)
	}

	hex := fmt.Sprintf("%X", int(rounded))
	if len(hex)%2 != 0 {
		hex = "0" + hex // 'A' to '0A'
	}

	return hex, nil
}

func parseIntRange(lower, upper, scale int, raw string) (string, error) {
	// expect a hex-representation of an integer value
	numeric, err := strconv.ParseInt(raw, 16, 64)
	if err != nil {
		return "", err
	}

	if scale == 0 {
		scale = 1
	}
	downscaled := float64(numeric) / float64(scale)

	// bounds check
	if downscaled < float64(lower) || downscaled > float64(upper) {
		return "", fmt.Errorf("invalid parameter %q", raw)
	}

	return fmt.Sprintf("%v", downscaled), nil
}

func formatIntRangeEnum(lower, upper, scale int, lookup map[string]string, raw interface{}) (string, error) {
	result, err := formatIntRange(lower, upper, scale, raw)
	if err == nil {
		return result, err
	}
	return formatEnum(lookup, raw)
}

func parseIntRangeEnum(lower, upper, scale int, lookup map[string]string, raw string) (string, error) {
	result, err := parseIntRange(lower, upper, scale, raw)
	if err == nil {
		return result, err
	}
	return parseEnum(lookup, raw)
}

func formatToggle(raw interface{}) (string, error) {
	s, ok := raw.(string)
	if ok {
		s = strings.ToLower(s)
		if s == "" || s == "toggle" || s == "tg" {
			return "TG", nil
		}
	}
	return "", fmt.Errorf("invalid parameter %q", raw)
}

func parseToggle(raw string) (string, error) {
	if raw == "TG" {
		return "toggle", nil
	}
	return "", fmt.Errorf("invalid parameter %q", raw)
}

// A CommandSet represents a set of known/supported commands
// and can be used to convert the "friendly" version to ISCP and vice-versa.
type CommandSet interface {
	// ReadCommand finds the command definition for an ISCP command
	// and converts the parameter.
	ReadCommand(command ISCPCommand) (string, string, error)

	// CreateCommand creates an ISCP command for the given friendly name
	// and parameter.
	// An error is returned if the name or parameter is invalid.
	CreateCommand(name string, param interface{}) (ISCPCommand, error)

	// CreateQuery creates a QSTN command for the given friendly name.
	CreateQuery(name string) (ISCPCommand, error)
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
		return "", "", fmt.Errorf("unknown ISCP command %q", command)
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
		return Command{}, fmt.Errorf("unknown command %q", name)
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
