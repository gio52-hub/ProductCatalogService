package domain

import (
	"math/big"
)

// Money represents a monetary value with precise decimal arithmetic using rational numbers.
// It stores values as numerator/denominator to avoid floating-point precision issues.
type Money struct {
	amount *big.Rat
}

// NewMoney creates a new Money instance from numerator and denominator.
// Example: NewMoney(1999, 100) represents $19.99
func NewMoney(numerator, denominator int64) *Money {
	if denominator == 0 {
		denominator = 1
	}
	return &Money{
		amount: big.NewRat(numerator, denominator),
	}
}

// NewMoneyFromRat creates a Money instance from an existing *big.Rat.
func NewMoneyFromRat(rat *big.Rat) *Money {
	if rat == nil {
		return &Money{amount: big.NewRat(0, 1)}
	}
	return &Money{amount: new(big.Rat).Set(rat)}
}

// Zero returns a Money instance representing zero.
func Zero() *Money {
	return &Money{amount: big.NewRat(0, 1)}
}

// Amount returns a copy of the underlying rational number.
func (m *Money) Amount() *big.Rat {
	if m == nil || m.amount == nil {
		return big.NewRat(0, 1)
	}
	return new(big.Rat).Set(m.amount)
}

// Numerator returns the numerator of the money value.
func (m *Money) Numerator() int64 {
	if m == nil || m.amount == nil {
		return 0
	}
	return m.amount.Num().Int64()
}

// Denominator returns the denominator of the money value.
func (m *Money) Denominator() int64 {
	if m == nil || m.amount == nil {
		return 1
	}
	return m.amount.Denom().Int64()
}

// Add returns a new Money that is the sum of m and other.
func (m *Money) Add(other *Money) *Money {
	if other == nil {
		return NewMoneyFromRat(m.Amount())
	}
	result := new(big.Rat).Add(m.Amount(), other.Amount())
	return NewMoneyFromRat(result)
}

// Sub returns a new Money that is the difference of m and other.
func (m *Money) Sub(other *Money) *Money {
	if other == nil {
		return NewMoneyFromRat(m.Amount())
	}
	result := new(big.Rat).Sub(m.Amount(), other.Amount())
	return NewMoneyFromRat(result)
}

// Multiply returns a new Money multiplied by the given rational number.
func (m *Money) Multiply(factor *big.Rat) *Money {
	if factor == nil {
		return NewMoneyFromRat(m.Amount())
	}
	result := new(big.Rat).Mul(m.Amount(), factor)
	return NewMoneyFromRat(result)
}

// CalculatePercentage returns a new Money representing the given percentage of m.
// percentage should be the percentage value (e.g., 20 for 20%).
func (m *Money) CalculatePercentage(percentage *big.Rat) *Money {
	if percentage == nil {
		return Zero()
	}
	// amount * (percentage / 100)
	factor := new(big.Rat).Quo(percentage, big.NewRat(100, 1))
	return m.Multiply(factor)
}

// ApplyDiscount returns a new Money after applying a percentage discount.
// percentage should be the discount percentage (e.g., 20 for 20% off).
func (m *Money) ApplyDiscount(percentage *big.Rat) *Money {
	if percentage == nil {
		return NewMoneyFromRat(m.Amount())
	}
	discountAmount := m.CalculatePercentage(percentage)
	return m.Sub(discountAmount)
}

// IsZero returns true if the money value is zero.
func (m *Money) IsZero() bool {
	if m == nil || m.amount == nil {
		return true
	}
	return m.amount.Sign() == 0
}

// IsPositive returns true if the money value is positive.
func (m *Money) IsPositive() bool {
	if m == nil || m.amount == nil {
		return false
	}
	return m.amount.Sign() > 0
}

// IsNegative returns true if the money value is negative.
func (m *Money) IsNegative() bool {
	if m == nil || m.amount == nil {
		return false
	}
	return m.amount.Sign() < 0
}

// Equals returns true if two Money values are equal.
func (m *Money) Equals(other *Money) bool {
	if m == nil && other == nil {
		return true
	}
	if m == nil || other == nil {
		return false
	}
	return m.Amount().Cmp(other.Amount()) == 0
}

// GreaterThan returns true if m is greater than other.
func (m *Money) GreaterThan(other *Money) bool {
	if m == nil || other == nil {
		return false
	}
	return m.Amount().Cmp(other.Amount()) > 0
}

// LessThan returns true if m is less than other.
func (m *Money) LessThan(other *Money) bool {
	if m == nil || other == nil {
		return false
	}
	return m.Amount().Cmp(other.Amount()) < 0
}

// String returns a string representation of the money value.
func (m *Money) String() string {
	if m == nil || m.amount == nil {
		return "0"
	}
	return m.amount.FloatString(2)
}
