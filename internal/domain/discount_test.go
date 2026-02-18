package domain

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDiscount_Valid(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	percentage := big.NewRat(20, 1)

	discount, err := NewDiscount(percentage, start, end)

	require.NoError(t, err)
	assert.NotNil(t, discount)
	assert.Equal(t, float64(20), discount.PercentageFloat())
	assert.Equal(t, start, discount.StartDate())
	assert.Equal(t, end, discount.EndDate())
}

func TestNewDiscount_InvalidPercentage(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		percentage *big.Rat
	}{
		{"nil percentage", nil},
		{"zero percentage", big.NewRat(0, 1)},
		{"negative percentage", big.NewRat(-10, 1)},
		{"over 100 percentage", big.NewRat(150, 1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDiscount(tt.percentage, start, end)
			assert.ErrorIs(t, err, ErrInvalidDiscountPercentage)
		})
	}
}

func TestNewDiscount_InvalidPeriod(t *testing.T) {
	percentage := big.NewRat(20, 1)

	tests := []struct {
		name  string
		start time.Time
		end   time.Time
	}{
		{
			name:  "end before start",
			start: time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "end equals start",
			start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDiscount(percentage, tt.start, tt.end)
			assert.ErrorIs(t, err, ErrInvalidDiscountPeriod)
		})
	}
}

func TestDiscount_IsValidAt(t *testing.T) {
	start := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
	percentage := big.NewRat(15, 1)

	discount, err := NewDiscount(percentage, start, end)
	require.NoError(t, err)

	tests := []struct {
		name     string
		checkAt  time.Time
		expected bool
	}{
		{
			name:     "before start",
			checkAt:  time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "at start",
			checkAt:  start,
			expected: true,
		},
		{
			name:     "during period",
			checkAt:  time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "at end (exclusive)",
			checkAt:  end,
			expected: false,
		},
		{
			name:     "after end",
			checkAt:  time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := discount.IsValidAt(tt.checkAt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDiscount_ApplyTo(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	percentage := big.NewRat(25, 1) // 25%

	discount, err := NewDiscount(percentage, start, end)
	require.NoError(t, err)

	basePrice := NewMoney(10000, 100) // $100.00
	discountedPrice := discount.ApplyTo(basePrice)

	expected := NewMoney(7500, 100) // $75.00 (25% off)
	assert.True(t, discountedPrice.Equals(expected))
}
