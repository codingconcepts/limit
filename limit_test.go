package limit

import (
	"testing"
	"time"

	"github.com/codingconcepts/limit/test"
)

func TestAllowed(t *testing.T) {
	rate := limit.Rate(2, time.Second)

	// Request at 0ms.
	allowed, left, err := rate.Allowed("test")
	test.ErrorNil(t, err)
	test.Equals(t, 2, left)
	test.Assert(t, allowed)

	// Request at 1000000ms.
	mockClock.Add(time.Millisecond)
	allowed, left, err = rate.Allowed("test")
	test.ErrorNil(t, err)
	test.Equals(t, 1, left)
	test.Assert(t, allowed)

	// Request at 2000000ms is disallowed because this is the third
	// request within 2 seconds.
	mockClock.Add(time.Millisecond)
	allowed, left, err = rate.Allowed("test")
	test.ErrorNil(t, err)
	test.Equals(t, 0, left)
	test.Assert(t, !allowed)

	// Request at 2002000000ms is allowed because this is outside
	// the 2 second window and hence part of the next window.
	mockClock.Add(time.Second * 2)
	allowed, left, err = rate.Allowed("test")
	test.ErrorNil(t, err)
	test.Equals(t, 2, left)
	test.Assert(t, allowed)
}
