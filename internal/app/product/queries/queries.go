package queries

import (
	"context"
	"time"

	"github.com/product-catalog-service/internal/app/product/contracts"
	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/product-catalog-service/internal/clock"
)

// GetProductRequest represents the input for getting a product.
type GetProductRequest struct {
	ProductID string
}

// ListProductsRequest represents the input for listing products.
type ListProductsRequest struct {
	Category   string
	Status     string
	ActiveOnly bool
	PageSize   int32
	PageToken  string
}

// ProductResponse represents the response for getting a product.
type ProductResponse struct {
	ID                        string
	Name                      string
	Description               string
	Category                  string
	BasePriceNumerator        int64
	BasePriceDenominator      int64
	EffectivePriceNumerator   int64
	EffectivePriceDenominator int64
	DiscountPercent           *float64
	DiscountStartDate         *time.Time
	DiscountEndDate           *time.Time
	HasActiveDiscount         bool
	Status                    string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

// ProductSummary represents a summary of a product in a list.
type ProductSummary struct {
	ID                        string
	Name                      string
	Category                  string
	BasePriceNumerator        int64
	BasePriceDenominator      int64
	EffectivePriceNumerator   int64
	EffectivePriceDenominator int64
	HasActiveDiscount         bool
	DiscountPercent           *float64
	Status                    string
	CreatedAt                 time.Time
}

// ListProductsResponse represents the response for listing products.
type ListProductsResponse struct {
	Products      []*ProductSummary
	NextPageToken string
	TotalCount    int64
}

// ProductQueries provides all product-related query operations.
type ProductQueries struct {
	readModel contracts.ProductReadModel
	clock     clock.Clock
}

// NewProductQueries creates a new ProductQueries instance.
func NewProductQueries(readModel contracts.ProductReadModel, clock clock.Clock) *ProductQueries {
	return &ProductQueries{
		readModel: readModel,
		clock:     clock,
	}
}

// GetProduct retrieves a product by ID with its current effective price.
func (q *ProductQueries) GetProduct(ctx context.Context, req GetProductRequest) (*ProductResponse, error) {
	if req.ProductID == "" {
		return nil, domain.ErrInvalidID
	}

	now := q.clock.Now()
	dto, err := q.readModel.GetProduct(ctx, req.ProductID, now)
	if err != nil {
		return nil, err
	}

	return productResponseFromDTO(dto), nil
}

// ListProducts lists products with optional filters and pagination.
func (q *ProductQueries) ListProducts(ctx context.Context, req ListProductsRequest) (*ListProductsResponse, error) {
	filter := contracts.ListProductsFilter{
		Category:   req.Category,
		Status:     req.Status,
		ActiveOnly: req.ActiveOnly,
	}

	pagination := contracts.Pagination{
		PageSize:  req.PageSize,
		PageToken: req.PageToken,
	}

	if pagination.PageSize <= 0 {
		pagination.PageSize = 20
	}
	if pagination.PageSize > 100 {
		pagination.PageSize = 100
	}

	now := q.clock.Now()
	result, err := q.readModel.ListProducts(ctx, filter, pagination, now)
	if err != nil {
		return nil, err
	}

	return listProductsResponseFromDTOs(result), nil
}

// ListProductsByCategory lists products in a specific category.
func (q *ProductQueries) ListProductsByCategory(ctx context.Context, category string, pageSize int32, pageToken string) (*ListProductsResponse, error) {
	pagination := contracts.Pagination{
		PageSize:  pageSize,
		PageToken: pageToken,
	}

	if pagination.PageSize <= 0 {
		pagination.PageSize = 20
	}
	if pagination.PageSize > 100 {
		pagination.PageSize = 100
	}

	now := q.clock.Now()
	result, err := q.readModel.ListByCategory(ctx, category, pagination, now)
	if err != nil {
		return nil, err
	}

	return listProductsResponseFromDTOs(result), nil
}

func productResponseFromDTO(dto *contracts.ProductDTO) *ProductResponse {
	if dto == nil {
		return nil
	}
	return &ProductResponse{
		ID:                        dto.ID,
		Name:                      dto.Name,
		Description:               dto.Description,
		Category:                  dto.Category,
		BasePriceNumerator:        dto.BasePriceNum,
		BasePriceDenominator:      dto.BasePriceDenom,
		EffectivePriceNumerator:   dto.EffectivePriceNum,
		EffectivePriceDenominator: dto.EffectivePriceDenom,
		DiscountPercent:           dto.DiscountPercent,
		DiscountStartDate:         dto.DiscountStartDate,
		DiscountEndDate:           dto.DiscountEndDate,
		HasActiveDiscount:         dto.HasActiveDiscount,
		Status:                    dto.Status,
		CreatedAt:                 dto.CreatedAt,
		UpdatedAt:                 dto.UpdatedAt,
	}
}

func listProductsResponseFromDTOs(result *contracts.ListProductsResult) *ListProductsResponse {
	if result == nil {
		return &ListProductsResponse{
			Products: make([]*ProductSummary, 0),
		}
	}

	products := make([]*ProductSummary, len(result.Products))
	for i, dto := range result.Products {
		products[i] = &ProductSummary{
			ID:                        dto.ID,
			Name:                      dto.Name,
			Category:                  dto.Category,
			BasePriceNumerator:        dto.BasePriceNum,
			BasePriceDenominator:      dto.BasePriceDenom,
			EffectivePriceNumerator:   dto.EffectivePriceNum,
			EffectivePriceDenominator: dto.EffectivePriceDenom,
			HasActiveDiscount:         dto.HasActiveDiscount,
			DiscountPercent:           dto.DiscountPercent,
			Status:                    dto.Status,
			CreatedAt:                 dto.CreatedAt,
		}
	}

	return &ListProductsResponse{
		Products:      products,
		NextPageToken: result.NextPageToken,
		TotalCount:    result.TotalCount,
	}
}
