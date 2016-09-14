/* See the file "LICENSE.txt" for the full license governing this code. */

// Wrapper type for all functions that depend on current time
// This is useful to use a mock clock implementation for testing

package main

import (
	"time"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now()
}
