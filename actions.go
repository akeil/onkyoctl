package onkyoctl

import (
    "errors"
    "strings"
    "strconv"
)

type ISCPGroup string

type ParamType string
const (
    paramOnOff ParamType = "onOff"
    paramOnOffToggle ParamType = "onOffToggle"
    // lookup
    // lookupToggle
)

func SplitISCP(command ISCPCommand) (ISCPGroup, string) {
    s := string(command)
    group := ISCPGroup(s[0:3])
    param := s[3:]

    return group, param
}

const (
    query = "QSTN"
)

type Command struct {
    Name string
    Title string
    Group ISCPGroup
    ParamType ParamType
}

func (c *Command) QueryCommand() ISCPCommand {
    return ISCPCommand(string(c.Group) + query)
}

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
        }else if find == "off" {
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
