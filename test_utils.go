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

func assertErr(t *testing.T, err error) {
	if err == nil {
		t.Log("Expected error, got none.")
		t.Fail()
	}
}

func assertNoErr(t *testing.T, err error) {
	if err != nil {
		t.Logf("Unexpected error: %v.", err)
		t.Fail()
	}
}
