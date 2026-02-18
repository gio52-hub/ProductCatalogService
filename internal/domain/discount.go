package domain

import (
	"math/big"
	"time"
)

// Discount represents a percentage-based discount with a validity period.
type Discount struct {
	percentage *big.Rat
	startDate  time.Time
	endDate    time.Time
}

// NewDiscount creates a new Discount value object.
// percentage is the discount percentage (e.g., 20 for 20% off).
func NewDiscount(percentage *big.Rat, startDate, endDate time.Time) (*Discount, error) {
	if percentage == nil {
		return nil, ErrInvalidDiscountPercentage
	}

	// Percentage must be between 0 and 100 (exclusive of 0, inclusive of 100)
	zero := big.NewRat(0, 1)
	hundred := big.NewRat(100, 1)

	if percentage.Cmp(zero) <= 0 {
		return nil, ErrInvalidDiscountPercentage
	}
	if percentage.Cmp(hundred) > 0 {
		return nil, ErrInvalidDiscountPercentage
	}

	// End date must be after start date
	if !endDate.After(startDate) {
		return nil, ErrInvalidDiscountPeriod
	}

	return &Discount{
		percentage: new(big.Rat).Set(percentage),
		startDate:  startDate,
		endDate:    endDate,
	}, nil
}

// Percentage returns a copy of the discount percentage.
func (d *Discount) Percentage() *big.Rat {
	if d == nil || d.percentage == nil {
		return big.NewRat(0, 1)
	}
	return new(big.Rat).Set(d.percentage)
}

// PercentageFloat returns the discount percentage as a float64.
func (d *Discount) PercentageFloat() float64 {
	if d == nil || d.percentage == nil {
		return 0
	}
	f, _ := d.percentage.Float64()
	return f
}

// StartDate returns the start date of the discount period.
func (d *Discount) StartDate() time.Time {
	if d == nil {
		return time.Time{}
	}
	return d.startDate
}

// EndDate returns the end date of the discount period.
func (d *Discount) EndDate() time.Time {
	if d == nil {
		return time.Time{}
	}
	return d.endDate
}

// IsValidAt checks if the discount is valid at the given time.
// A discount is valid if the time is within the start and end dates (inclusive of start, exclusive of end).
func (d *Discount) IsValidAt(t time.Time) bool {
	if d == nil {
		return false
	}
	return !t.Before(d.startDate) && t.Before(d.endDate)
}

// IsActive is an alias for IsValidAt.
func (d *Discount) IsActive(t time.Time) bool {
	return d.IsValidAt(t)
}

// IsExpired checks if the discount has expired at the given time.
func (d *Discount) IsExpired(t time.Time) bool {
	if d == nil {
		return true
	}
	return !t.Before(d.endDate)
}

// HasStarted checks if the discount period has started at the given time.
func (d *Discount) HasStarted(t time.Time) bool {
	if d == nil {
		return false
	}
	return !t.Before(d.startDate)
}

// ApplyTo calculates the discounted price for a given Money value.
func (d *Discount) ApplyTo(price *Money) *Money {
	if d == nil || price == nil {
		return price
	}
	return price.ApplyDiscount(d.percentage)
}

// Equals checks if two discounts are equal.
func (d *Discount) Equals(other *Discount) bool {
	if d == nil && other == nil {
		return true
	}
	if d == nil || other == nil {
		return false
	}
	return d.percentage.Cmp(other.percentage) == 0 &&
		d.startDate.Equal(other.startDate) &&
		d.endDate.Equal(other.endDate)
}
