package test

import (
	"reflect"
	"testing"
)

// Equals performs a deep equal comparison against two
// values and fails if they are not the same.
func Equals(tb testing.TB, expected, actual interface{}) {
	tb.Helper()
	if !reflect.DeepEqual(expected, actual) {
		tb.Fatalf("expected: %#[1]v (%[1]T) but got: %#[2]v (%[2]T)\n", expected, actual)
	}
}

// Assert asserts that a condition is met and fails if it's not.
func Assert(tb testing.TB, result bool) {
	tb.Helper()
	if !result {
		tb.Fatal("assertion failed\n")
	}
}

// ErrorNil asserts that an error is nil and fails if it's not.
func ErrorNil(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
}
