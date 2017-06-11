package test

import "testing"

func Asserte(t *testing.T, b bool, msg string, args ...interface{}) {
	if !b {
		t.Errorf(msg, args...)
	}
}
