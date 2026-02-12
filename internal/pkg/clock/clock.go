package clock

import (
	"time"
)

// Clock provides time abstraction
type Clock interface {
	Now() time.Time
}

// RealClock implements Clock using the actual system time
type RealClock struct{}

// NewRealClock creates a new real clock
func NewRealClock() *RealClock {
	return &RealClock{}
}

// Now returns the current time
func (c *RealClock) Now() time.Time {
	return time.Now().UTC()
}

// MockClock implements Clock with a fixed time for testing
type MockClock struct {
	FixedTime time.Time
}

// NewMockClock creates a new mock clock with a fixed time
func NewMockClock(t time.Time) *MockClock {
	return &MockClock{FixedTime: t}
}

// Now returns the fixed time
func (c *MockClock) Now() time.Time {
	return c.FixedTime
}
