package test_utils

import "testing"

func AssertEqual(t *testing.T, got, expected interface{}) {
	t.Helper()

	if got != expected {
		t.Errorf("got %v want %v", got, expected)
	}
}
