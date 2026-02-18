package domain

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name        string
		numerator   int64
		denominator int64
		wantNum     int64
		wantDenom   int64
	}{
		{
			name:        "standard price $19.99",
			numerator:   1999,
			denominator: 100,
			wantNum:     1999,
			wantDenom:   100,
		},
		{
			name:        "whole dollar $50.00",
			numerator:   5000,
			denominator: 100,
			wantNum:     50,
			wantDenom:   1,
		},
		{
			name:        "zero denominator defaults to 1",
			numerator:   100,
			denominator: 0,
			wantNum:     100,
			wantDenom:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMoney(tt.numerator, tt.denominator)
			// Check that the ratio is equivalent
			expected := big.NewRat(tt.wantNum, tt.wantDenom)
			assert.Equal(t, 0, m.Amount().Cmp(expected), "amounts should be equal")
		})
	}
}

func TestMoney_Add(t *testing.T) {
	m1 := NewMoney(1999, 100) // $19.99
	m2 := NewMoney(500, 100)  // $5.00

	result := m1.Add(m2)

	expected := NewMoney(2499, 100) // $24.99
	assert.True(t, result.Equals(expected))
}

func TestMoney_Sub(t *testing.T) {
	m1 := NewMoney(2000, 100) // $20.00
	m2 := NewMoney(500, 100)  // $5.00

	result := m1.Sub(m2)

	expected := NewMoney(1500, 100) // $15.00
	assert.True(t, result.Equals(expected))
}

func TestMoney_CalculatePercentage(t *testing.T) {
	base := NewMoney(10000, 100) // $100.00

	tests := []struct {
		name       string
		percentage *big.Rat
		wantNum    int64
		wantDenom  int64
	}{
		{
			name:       "10%",
			percentage: big.NewRat(10, 1),
			wantNum:    1000,
			wantDenom:  100,
		},
		{
			name:       "20%",
			percentage: big.NewRat(20, 1),
			wantNum:    2000,
			wantDenom:  100,
		},
		{
			name:       "50%",
			percentage: big.NewRat(50, 1),
			wantNum:    5000,
			wantDenom:  100,
		},
		{
			name:       "15.5%",
			percentage: big.NewRat(155, 10),
			wantNum:    1550,
			wantDenom:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := base.CalculatePercentage(tt.percentage)
			expected := NewMoney(tt.wantNum, tt.wantDenom)
			assert.True(t, result.Equals(expected), "got %s, want %s", result.String(), expected.String())
		})
	}
}

func TestMoney_ApplyDiscount(t *testing.T) {
	base := NewMoney(10000, 100) // $100.00

	tests := []struct {
		name       string
		discount   *big.Rat
		wantNum    int64
		wantDenom  int64
	}{
		{
			name:      "10% off",
			discount:  big.NewRat(10, 1),
			wantNum:   9000,
			wantDenom: 100,
		},
		{
			name:      "20% off",
			discount:  big.NewRat(20, 1),
			wantNum:   8000,
			wantDenom: 100,
		},
		{
			name:      "25% off",
			discount:  big.NewRat(25, 1),
			wantNum:   7500,
			wantDenom: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := base.ApplyDiscount(tt.discount)
			expected := NewMoney(tt.wantNum, tt.wantDenom)
			assert.True(t, result.Equals(expected), "got %s, want %s", result.String(), expected.String())
		})
	}
}

func TestMoney_Comparisons(t *testing.T) {
	m1 := NewMoney(1000, 100)
	m2 := NewMoney(500, 100)
	m3 := NewMoney(1000, 100)

	assert.True(t, m1.GreaterThan(m2))
	assert.True(t, m2.LessThan(m1))
	assert.True(t, m1.Equals(m3))
	assert.False(t, m1.IsZero())
	assert.True(t, m1.IsPositive())
	assert.False(t, m1.IsNegative())
}

func TestMoney_Zero(t *testing.T) {
	z := Zero()
	assert.True(t, z.IsZero())
	assert.False(t, z.IsPositive())
	assert.False(t, z.IsNegative())
}
