// Package clock provides a time abstraction for deterministic testing.
package clock

import "time"

// Clock is an interface for getting the current time.
// This abstraction allows for easy testing with fixed times.
type Clock interface {
	Now() time.Time
}

// RealClock implements Clock using the system clock.
type RealClock struct{}

// NewRealClock creates a new RealClock instance.
func NewRealClock() *RealClock {
	return &RealClock{}
}

// Now returns the current system time.
func (c *RealClock) Now() time.Time {
	return time.Now()
}

// FixedClock implements Clock with a fixed time for testing.
type FixedClock struct {
	fixedTime time.Time
}

// NewFixedClock creates a new FixedClock with the given time.
func NewFixedClock(t time.Time) *FixedClock {
	return &FixedClock{fixedTime: t}
}

// Now returns the fixed time.
func (c *FixedClock) Now() time.Time {
	return c.fixedTime
}

// SetTime updates the fixed time.
func (c *FixedClock) SetTime(t time.Time) {
	c.fixedTime = t
}

// Advance moves the fixed time forward by the given duration.
func (c *FixedClock) Advance(d time.Duration) {
	c.fixedTime = c.fixedTime.Add(d)
}
