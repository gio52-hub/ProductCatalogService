package domain

import "errors"

// Domain errors are sentinel values that represent business rule violations.
var (
	// Product errors
	ErrProductNotFound    = errors.New("product not found")
	ErrProductNotActive   = errors.New("product is not active")
	ErrProductArchived    = errors.New("product is archived")
	ErrProductAlreadyActive = errors.New("product is already active")
	ErrProductAlreadyInactive = errors.New("product is already inactive")
	ErrInvalidProductName = errors.New("invalid product name")
	ErrInvalidProductCategory = errors.New("invalid product category")
	ErrInvalidBasePrice   = errors.New("base price must be positive")

	// Discount errors
	ErrInvalidDiscountPercentage = errors.New("discount percentage must be between 0 and 100")
	ErrInvalidDiscountPeriod     = errors.New("discount end date must be after start date")
	ErrDiscountNotActive         = errors.New("discount is not active at the current time")
	ErrDiscountAlreadyExists     = errors.New("product already has an active discount")
	ErrNoDiscountToRemove        = errors.New("product has no discount to remove")

	// General errors
	ErrInvalidID = errors.New("invalid ID")
)
