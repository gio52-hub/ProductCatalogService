package product

import (
	"errors"

	pb "github.com/product-catalog-service/proto/product/v1"
)

// Validation errors
var (
	ErrProductIDRequired      = errors.New("product_id is required")
	ErrNameRequired           = errors.New("name is required")
	ErrCategoryRequired       = errors.New("category is required")
	ErrBasePriceRequired      = errors.New("base_price is required")
	ErrInvalidBasePrice       = errors.New("base_price must be positive")
	ErrDiscountRequired       = errors.New("discount_percentage is required")
	ErrInvalidDiscount        = errors.New("discount_percentage must be between 0 and 100")
	ErrStartDateRequired      = errors.New("start_date is required")
	ErrEndDateRequired        = errors.New("end_date is required")
	ErrEndDateBeforeStartDate = errors.New("end_date must be after start_date")
)

// validateCreateRequest validates a CreateProductRequest.
func validateCreateRequest(req *pb.CreateProductRequest) error {
	if req.GetName() == "" {
		return ErrNameRequired
	}
	if req.GetCategory() == "" {
		return ErrCategoryRequired
	}
	if req.GetBasePrice() == nil {
		return ErrBasePriceRequired
	}
	if req.GetBasePrice().GetNumerator() <= 0 || req.GetBasePrice().GetDenominator() <= 0 {
		return ErrInvalidBasePrice
	}
	return nil
}

// validateUpdateRequest validates an UpdateProductRequest.
func validateUpdateRequest(req *pb.UpdateProductRequest) error {
	if req.GetProductId() == "" {
		return ErrProductIDRequired
	}
	if req.GetName() == "" {
		return ErrNameRequired
	}
	if req.GetCategory() == "" {
		return ErrCategoryRequired
	}
	return nil
}

// validateApplyDiscountRequest validates an ApplyDiscountRequest.
func validateApplyDiscountRequest(req *pb.ApplyDiscountRequest) error {
	if req.GetProductId() == "" {
		return ErrProductIDRequired
	}
	if req.GetDiscountPercentage() <= 0 || req.GetDiscountPercentage() > 100 {
		return ErrInvalidDiscount
	}
	if req.GetStartDate() == nil {
		return ErrStartDateRequired
	}
	if req.GetEndDate() == nil {
		return ErrEndDateRequired
	}
	if !req.GetEndDate().AsTime().After(req.GetStartDate().AsTime()) {
		return ErrEndDateBeforeStartDate
	}
	return nil
}
