package onkyoctl

import (
    "reflect"
	"testing"
)

func assertEqual(t *testing.T, actual, expected interface{}) {
	assert(t, actual, expected, true)
}

func assert(t *testing.T, actual, expected interface{}, shouldEqual bool) {
	if reflect.DeepEqual(expected, actual) != shouldEqual {
		t.Logf("Expected %q, got %q.", expected, actual)
		t.Fail()
	}
}
