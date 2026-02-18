package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealClock_Now(t *testing.T) {
	t.Parallel()

	c := NewRealClock()
	before := time.Now()
	actual := c.Now()
	after := time.Now()

	assert.True(t, actual.After(before) || actual.Equal(before), "Clock time should be >= before")
	assert.True(t, actual.Before(after) || actual.Equal(after), "Clock time should be <= after")
}

func TestFixedClock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setTime  time.Time
		wantTime time.Time
	}{
		{
			name:     "zero time",
			setTime:  time.Time{},
			wantTime: time.Time{},
		},
		{
			name:     "specific time",
			setTime:  time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
			wantTime: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:     "unix epoch",
			setTime:  time.Unix(0, 0),
			wantTime: time.Unix(0, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewFixedClock(tt.setTime)
			require.NotNil(t, c)
			assert.Equal(t, tt.wantTime, c.Now())
		})
	}
}

func TestFixedClock_SetTime(t *testing.T) {
	t.Parallel()

	c := NewFixedClock(time.Time{})

	firstTime := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	c.SetTime(firstTime)
	assert.Equal(t, firstTime, c.Now())

	secondTime := time.Date(2026, 12, 25, 8, 30, 0, 0, time.UTC)
	c.SetTime(secondTime)
	assert.Equal(t, secondTime, c.Now())
}

func TestFixedClock_Advance(t *testing.T) {
	t.Parallel()

	startTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	c := NewFixedClock(startTime)

	tests := []struct {
		name        string
		advanceBy   time.Duration
		expectedGap time.Duration
	}{
		{
			name:        "advance by 1 hour",
			advanceBy:   1 * time.Hour,
			expectedGap: 1 * time.Hour,
		},
		{
			name:        "advance by 24 hours",
			advanceBy:   24 * time.Hour,
			expectedGap: 25 * time.Hour,
		},
		{
			name:        "advance by 1 minute",
			advanceBy:   1 * time.Minute,
			expectedGap: 25*time.Hour + 1*time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.Advance(tt.advanceBy)
			assert.Equal(t, startTime.Add(tt.expectedGap), c.Now())
		})
	}
}

func TestClock_Interface(t *testing.T) {
	t.Parallel()

	var _ Clock = NewRealClock()
	var _ Clock = NewFixedClock(time.Now())
}
