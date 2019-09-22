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

    query := c.QueryCommand()
    assertEqual(t, query, ISCPCommand("PWRQSTN"))
}

func TestFormatOnOff(t *testing.T) {
    c := &Command{
        Group: "PWR",
        ParamType: "onOff",
    }

    type TestCase struct {
        Param interface{}
        ExpectError bool
        Expected ISCPCommand
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
    for _, tc := range(cases) {
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
