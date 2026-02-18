package contract

import (
	"context"
	"time"
)

// ProductDTO represents a product for read operations.
type ProductDTO struct {
	ID                 string
	Name               string
	Description        string
	Category           string
	BasePriceNum       int64
	BasePriceDenom     int64
	DiscountPercent    *float64
	DiscountStartDate  *time.Time
	DiscountEndDate    *time.Time
	EffectivePriceNum  int64
	EffectivePriceDenom int64
	Status             string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	HasActiveDiscount  bool
}

// ListProductsFilter defines filters for listing products.
type ListProductsFilter struct {
	Category   string
	Status     string
	ActiveOnly bool
}

// Pagination defines pagination parameters.
type Pagination struct {
	PageSize  int32
	PageToken string
}

// ListProductsResult represents the result of listing products.
type ListProductsResult struct {
	Products      []*ProductDTO
	NextPageToken string
	TotalCount    int64
}

// ProductReadModel defines the interface for product read operations (queries).
// Following CQRS, queries bypass the domain layer for optimization.
type ProductReadModel interface {
	// GetProduct retrieves a product by ID with its current effective price.
	GetProduct(ctx context.Context, id string, at time.Time) (*ProductDTO, error)

	// ListProducts lists products with optional filters and pagination.
	ListProducts(ctx context.Context, filter ListProductsFilter, pagination Pagination, at time.Time) (*ListProductsResult, error)

	// ListByCategory lists products in a specific category.
	ListByCategory(ctx context.Context, category string, pagination Pagination, at time.Time) (*ListProductsResult, error)

	// CountByCategory returns the count of active products in a category.
	CountByCategory(ctx context.Context, category string) (int64, error)
}
