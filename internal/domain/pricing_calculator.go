package domain

import (
	"math/big"
	"time"
)

// PricingCalculator is a domain service for pricing calculations.
type PricingCalculator struct{}

// NewPricingCalculator creates a new PricingCalculator instance.
func NewPricingCalculator() *PricingCalculator {
	return &PricingCalculator{}
}

// CalculateEffectivePrice calculates the effective price for a product at a given time.
func (pc *PricingCalculator) CalculateEffectivePrice(product *Product, at time.Time) *Money {
	if product == nil {
		return Zero()
	}
	return product.EffectivePrice(at)
}

// CalculateDiscountedPrice calculates the price after applying a discount percentage.
func (pc *PricingCalculator) CalculateDiscountedPrice(basePrice *Money, discountPercent *big.Rat) *Money {
	if basePrice == nil {
		return Zero()
	}
	if discountPercent == nil {
		return basePrice
	}
	return basePrice.ApplyDiscount(discountPercent)
}

// CalculateDiscountAmount calculates the discount amount for a given base price and percentage.
func (pc *PricingCalculator) CalculateDiscountAmount(basePrice *Money, discountPercent *big.Rat) *Money {
	if basePrice == nil || discountPercent == nil {
		return Zero()
	}
	return basePrice.CalculatePercentage(discountPercent)
}

// CalculateSavings calculates how much a customer saves with the current discount.
func (pc *PricingCalculator) CalculateSavings(product *Product, at time.Time) *Money {
	if product == nil {
		return Zero()
	}
	if !product.HasActiveDiscount(at) {
		return Zero()
	}
	discount := product.Discount()
	if discount == nil {
		return Zero()
	}
	return pc.CalculateDiscountAmount(product.BasePrice(), discount.Percentage())
}
