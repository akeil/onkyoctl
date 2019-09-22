package onkyoctl

import (
	"testing"
)

func TestISCPSplit(t *testing.T) {
	var command ISCPCommand
	command = "PWR01"

	group, param := SplitISCP(command)
	assertEqual(t, group, ISCPGroup("PWR"))
	assertEqual(t, param, "01")
}

func TestFriendlyGenerateQuery(t *testing.T) {
	c := &Command{
		Group: "PWR",
	}

	query := c.CreateQuery()
	assertEqual(t, query, ISCPCommand("PWRQSTN"))
}

func TestFormatOnOff(t *testing.T) {
	c := &Command{
		Group:     "PWR",
		ParamType: "onOff",
	}

	type TestCase struct {
		Param       interface{}
		ExpectError bool
		Expected    ISCPCommand
	}
	cases := []TestCase{
		// booleans
		TestCase{
			Param: true, ExpectError: false, Expected: ISCPCommand("PWR01"),
		},
		TestCase{
			Param: false, ExpectError: false, Expected: ISCPCommand("PWR00"),
		},
		// numeric
		TestCase{
			Param: 1, ExpectError: false, Expected: ISCPCommand("PWR01"),
		},
		TestCase{
			Param: 0, ExpectError: false, Expected: ISCPCommand("PWR00"),
		},
		TestCase{
			Param: 1.0, ExpectError: false, Expected: ISCPCommand("PWR01"),
		},
		TestCase{
			Param: 0.0, ExpectError: false, Expected: ISCPCommand("PWR00"),
		},
		TestCase{
			Param: 2, ExpectError: true,
		},
		TestCase{
			Param: 0.5, ExpectError: true,
		},
		// strings
		TestCase{
			Param: "on", ExpectError: false, Expected: ISCPCommand("PWR01"),
		},
		TestCase{
			Param: "ON", ExpectError: false, Expected: ISCPCommand("PWR01"),
		},
		TestCase{
			Param: "true", ExpectError: false, Expected: ISCPCommand("PWR01"),
		},
		TestCase{
			Param: "TRUE", ExpectError: false, Expected: ISCPCommand("PWR01"),
		},
		TestCase{
			Param: "1", ExpectError: false, Expected: ISCPCommand("PWR01"),
		},
		TestCase{
			Param: "01", ExpectError: false, Expected: ISCPCommand("PWR01"),
		},
		TestCase{
			Param: "off", ExpectError: false, Expected: ISCPCommand("PWR00"),
		},
		TestCase{
			Param: "OFF", ExpectError: false, Expected: ISCPCommand("PWR00"),
		},
		TestCase{
			Param: "false", ExpectError: false, Expected: ISCPCommand("PWR00"),
		},
		TestCase{
			Param: "FALSE", ExpectError: false, Expected: ISCPCommand("PWR00"),
		},
		TestCase{
			Param: "0", ExpectError: false, Expected: ISCPCommand("PWR00"),
		},
		TestCase{
			Param: "00", ExpectError: false, Expected: ISCPCommand("PWR00"),
		},
		// invalid strings
		TestCase{
			Param: "foo", ExpectError: true,
		},
		TestCase{
			Param: "X", ExpectError: true,
		},
	}
	for _, tc := range cases {
		actual, err := c.CreateCommand(tc.Param)
		if tc.ExpectError {
			assertErr(t, err)
		} else {
			assertEqual(t, actual, tc.Expected)
		}
	}

	var err error
	var actual ISCPCommand
	// toggle not allowed
	_, err = c.CreateCommand("toggle")
	assertErr(t, err)
	_, err = c.CreateCommand("")
	assertErr(t, err)

	// including toggle
	c.ParamType = "onOffToggle"
	toggle := ISCPCommand("PWRTG")
	actual, err = c.CreateCommand("toggle")
	assertNoErr(t, err)
	assertEqual(t, actual, toggle)
	actual, err = c.CreateCommand("TOGGLE")
	assertNoErr(t, err)
	assertEqual(t, actual, toggle)
	actual, err = c.CreateCommand("tg")
	assertNoErr(t, err)
	assertEqual(t, actual, toggle)
	actual, err = c.CreateCommand("TG")
	assertNoErr(t, err)
	assertEqual(t, actual, toggle)
	actual, err = c.CreateCommand("")
	assertNoErr(t, err)
	assertEqual(t, actual, toggle)
}

func TestParseOnOff(t *testing.T) {
	c := Command{
		Group:     "PWR",
		ParamType: "onOff",
	}
	type TestCase struct {
		Raw         string
		ExpectError bool
		Expected    string
	}
	cases := []TestCase{
		TestCase{Raw: "01", ExpectError: false, Expected: "on"},
		TestCase{Raw: "00", ExpectError: false, Expected: "off"},
		TestCase{Raw: "xx", ExpectError: true},
		TestCase{Raw: "", ExpectError: true},
	}

	var actual string
	var err error
	for _, tc := range cases {
		actual, err = c.ParseParam(tc.Raw)
		if tc.ExpectError {
			assertErr(t, err)
		} else {
			assertEqual(t, actual, tc.Expected)
		}
	}

	// no toggle
	_, err = c.ParseParam("TG")
	assertErr(t, err)

	// with toggle
	c.ParamType = "onOffToggle"
	actual, err = c.ParseParam("TG")
	assertNoErr(t, err)
	assertEqual(t, actual, "toggle")
}

func TestFormatEnum(t *testing.T) {
	c := Command{
		Group:     "DIM",
		ParamType: "enum",
		Lookup: map[string]string{
			"00": "bright",
			"01": "dim",
			"02": "dark",
			"03": "off",
			"08": "led-off",
		},
	}

	var err error
	var actual ISCPCommand

	actual, err = c.CreateCommand("bright")
	assertNoErr(t, err)
	assertEqual(t, actual, ISCPCommand("DIM00"))

	actual, err = c.CreateCommand("off")
	assertNoErr(t, err)
	assertEqual(t, actual, ISCPCommand("DIM03"))

	actual, err = c.CreateCommand("Off")
	assertNoErr(t, err)
	assertEqual(t, actual, ISCPCommand("DIM03"))

	_, err = c.CreateCommand("unknown")
	assertErr(t, err)

	_, err = c.CreateCommand("")
	assertErr(t, err)

	_, err = c.CreateCommand(123)
	assertErr(t, err)

	_, err = c.CreateCommand(true)
	assertErr(t, err)

	_, err = c.CreateCommand("toggle")
	assertErr(t, err)

	c.ParamType = "enumToggle"
	actual, err = c.CreateCommand("toggle")
	assertNoErr(t, err)
	assertEqual(t, actual, ISCPCommand("DIMTG"))
}

func TestParseEnum(t *testing.T) {
	c := Command{
		Group:     "DIM",
		ParamType: "enum",
		Lookup: map[string]string{
			"00": "bright",
			"01": "dim",
			"02": "dark",
			"03": "off",
			"08": "led-off",
		},
	}

	var err error
	var actual string

	actual, err = c.ParseParam("03")
	assertNoErr(t, err)
	assertEqual(t, actual, "off")

	actual, err = c.ParseParam("08")
	assertNoErr(t, err)
	assertEqual(t, actual, "led-off")

	_, err = c.ParseParam("invalid")
	assertErr(t, err)

	_, err = c.ParseParam("123")
	assertErr(t, err)

	_, err = c.ParseParam("")
	assertErr(t, err)

	c.ParamType = "enumToggle"
	actual, err = c.ParseParam("TG")
	assertNoErr(t, err)
	assertEqual(t, actual, "toggle")

	actual, err = c.ParseParam("00")
	assertNoErr(t, err)
	assertEqual(t, actual, "bright")
}

func TestFormatIntRange(t *testing.T) {
	c := Command{
		Group:     "MVL",
		ParamType: "intRangeEnum",
		Lower:     0,
		Upper:     100,
		Scale:     2,
		Lookup: map[string]string{
			"UP":   "up",
			"DOWN": "down",
		},
	}

	var err error
	var actual ISCPCommand

	actual, err = c.CreateCommand(23) // x2 = 46 / 0x2e
	assertNoErr(t, err)
	assertEqual(t, actual, ISCPCommand("MVL2E"))

	actual, err = c.CreateCommand(23.0) // x2 = 46 / 0x2e
	assertNoErr(t, err)
	assertEqual(t, actual, ISCPCommand("MVL2E"))

	actual, err = c.CreateCommand(2.5) // x2 = 5 / 0x5
	assertNoErr(t, err)
	assertEqual(t, actual, ISCPCommand("MVL05"))

	actual, err = c.CreateCommand(0)
	assertNoErr(t, err)
	assertEqual(t, actual, ISCPCommand("MVL00"))

	// parse from string
	actual, err = c.CreateCommand("23.0")
	assertNoErr(t, err)
	assertEqual(t, actual, ISCPCommand("MVL2E"))

	actual, err = c.CreateCommand("2.5")
	assertNoErr(t, err)
	assertEqual(t, actual, ISCPCommand("MVL05"))

	// enum entries
	actual, err = c.CreateCommand("up")
	assertNoErr(t, err)
	assertEqual(t, actual, ISCPCommand("MVLUP"))

	// out of range
	_, err = c.CreateCommand(105)
	assertErr(t, err)

	_, err = c.CreateCommand(100.1)
	assertErr(t, err)

	_, err = c.CreateCommand(-1)
	assertErr(t, err)

	// rounding-error-guard
	_, err = c.CreateCommand(11.3)
	assertErr(t, err)

	// type
	_, err = c.CreateCommand(true)
	assertErr(t, err)
	_, err = c.CreateCommand("abc")
	assertErr(t, err)
	_, err = c.CreateCommand("")
	assertErr(t, err)
}

func TestParseIntRange(t *testing.T) {
	c := Command{
		Group:     "MVL",
		ParamType: "intRangeEnum",
		Lower:     0,
		Upper:     100,
		Scale:     2,
		Lookup: map[string]string{
			"UP":   "up",
			"DOWN": "down",
		},
	}

	var err error
	var actual string

	actual, err = c.ParseParam("00")
	assertNoErr(t, err)
	assertEqual(t, actual, "0")

	actual, err = c.ParseParam("05")
	assertNoErr(t, err)
	assertEqual(t, actual, "2.5")

	actual, err = c.ParseParam("2E")
	assertNoErr(t, err)
	assertEqual(t, actual, "23")

	// enum
	actual, err = c.ParseParam("DOWN")
	assertNoErr(t, err)
	assertEqual(t, actual, "down")

	// not a number
	_, err = c.ParseParam("XX")
	assertErr(t, err)

	_, err = c.ParseParam("")
	assertErr(t, err)

	// out of range
	_, err = c.ParseParam("FF")
	assertErr(t, err)
}

func TestBasicCreate(t *testing.T) {
	commands := []Command{
		Command{
			Name:      "power",
			Group:     "PWR",
			ParamType: "onOff",
		},
		Command{
			Name:      "mute",
			Group:     "AMT",
			ParamType: "onOffToggle",
		},
	}
	cs := NewBasicCommandSet(commands)

	type TestCase struct {
		Name        string
		Param       interface{}
		Expected    ISCPCommand
		ExpectError bool
	}

	cases := []TestCase{
		TestCase{
			Name:        "power",
			Param:       "on",
			Expected:    ISCPCommand("PWR01"),
			ExpectError: false,
		},
		TestCase{
			Name:        "power",
			Param:       "Off",
			Expected:    ISCPCommand("PWR00"),
			ExpectError: false,
		},
		TestCase{
			Name:        "mute",
			Param:       "toggle",
			Expected:    ISCPCommand("AMTTG"),
			ExpectError: false,
		},
		// unsupported param
		TestCase{
			Name:        "power",
			Param:       "toggle",
			ExpectError: true,
		},
		// unsupported command name
		TestCase{
			Name:        "unknown",
			Param:       "on",
			ExpectError: true,
		},
	}

	for _, tc := range cases {
		actual, err := cs.CreateCommand(tc.Name, tc.Param)
		if tc.ExpectError {
			assertErr(t, err)
		} else {
			assertEqual(t, actual, tc.Expected)
		}
	}
}

func TestBasicRead(t *testing.T) {
	commands := []Command{
		Command{
			Name:      "power",
			Group:     "PWR",
			ParamType: "onOff",
		},
		Command{
			Name:      "mute",
			Group:     "AMT",
			ParamType: "onOffToggle",
		},
	}
	cs := NewBasicCommandSet(commands)

	type TestCase struct {
		ISCP        ISCPCommand
		ExpectError bool
		Name        string
		Value       string
	}
	cases := []TestCase{
		TestCase{ISCP: "PWR01", ExpectError: false, Name: "power", Value: "on"},
		TestCase{ISCP: "PWR00", ExpectError: false, Name: "power", Value: "off"},
		TestCase{ISCP: "PWRxx", ExpectError: true},
		TestCase{ISCP: "PWR", ExpectError: true},

		TestCase{ISCP: "AMTTG", ExpectError: false, Name: "mute", Value: "toggle"},

		TestCase{ISCP: "FOO", ExpectError: true},
	}

	for _, tc := range cases {
		name, value, err := cs.ReadCommand(tc.ISCP)
		if tc.ExpectError {
			assertErr(t, err)
		} else {
			assertEqual(t, name, tc.Name)
			assertEqual(t, value, tc.Value)
		}
	}

}

func TestBasicQuery(t *testing.T) {
	commands := []Command{
		Command{
			Name:      "power",
			Group:     "PWR",
			ParamType: "onOff",
		},
		Command{
			Name:      "mute",
			Group:     "AMT",
			ParamType: "onOffToggle",
		},
	}
	cs := NewBasicCommandSet(commands)

	var actual ISCPCommand
	var err error

	actual, err = cs.CreateQuery("power")
	assertNoErr(t, err)
	assertEqual(t, actual, ISCPCommand("PWRQSTN"))

	actual, err = cs.CreateQuery("mute")
	assertNoErr(t, err)
	assertEqual(t, actual, ISCPCommand("AMTQSTN"))

	_, err = cs.CreateQuery("unkown")
	assertErr(t, err)
}
