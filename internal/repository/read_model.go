package repository

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/product-catalog-service/internal/contract"
	"github.com/product-catalog-service/internal/domain"
	"google.golang.org/api/iterator"
)

// ProductReadModel implements the contract.ProductReadModel interface using Spanner.
type ProductReadModel struct {
	client *spanner.Client
}

// NewProductReadModel creates a new ProductReadModel.
func NewProductReadModel(client *spanner.Client) *ProductReadModel {
	return &ProductReadModel{client: client}
}

// GetProduct retrieves a product by ID with its current effective price.
func (rm *ProductReadModel) GetProduct(ctx context.Context, id string, at time.Time) (*contract.ProductDTO, error) {
	row, err := rm.client.Single().ReadRow(
		ctx,
ProductsTable,
		spanner.Key{id},
ProductAllColumns(),
	)
	if err != nil {
		if spanner.ErrCode(err) == 5 { // NOT_FOUND
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}

	return rm.rowToDTO(row, at)
}

// ListProducts lists products with optional filters and pagination.
func (rm *ProductReadModel) ListProducts(ctx context.Context, filter contract.ListProductsFilter, pagination contract.Pagination, at time.Time) (*contract.ListProductsResult, error) {
	stmt := rm.buildListQuery(filter, pagination)
	iter := rm.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	products := make([]*contract.ProductDTO, 0)
	var lastProductID string

	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		dto, err := rm.rowToDTO(row, at)
		if err != nil {
			return nil, err
		}

		products = append(products, dto)
		lastProductID = dto.ID
	}

	// Determine next page token
	var nextPageToken string
	if len(products) == int(pagination.PageSize) && lastProductID != "" {
		nextPageToken = lastProductID
	}

	return &contract.ListProductsResult{
		Products:      products,
		NextPageToken: nextPageToken,
	}, nil
}

// ListByCategory lists products in a specific category.
func (rm *ProductReadModel) ListByCategory(ctx context.Context, category string, pagination contract.Pagination, at time.Time) (*contract.ListProductsResult, error) {
	filter := contract.ListProductsFilter{
		Category:   category,
		ActiveOnly: true,
	}
	return rm.ListProducts(ctx, filter, pagination, at)
}

// CountByCategory returns the count of active products in a category.
func (rm *ProductReadModel) CountByCategory(ctx context.Context, category string) (int64, error) {
	stmt := spanner.Statement{
		SQL: `SELECT COUNT(*) as count FROM products WHERE category = @category AND status = @status`,
		Params: map[string]interface{}{
			"category": category,
			"status":   string(domain.ProductStatusActive),
		},
	}

	iter := rm.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err != nil {
		return 0, err
	}

	var count int64
	if err := row.Columns(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// buildListQuery builds the SQL query for listing products.
func (rm *ProductReadModel) buildListQuery(filter contract.ListProductsFilter, pagination contract.Pagination) spanner.Statement {
	sql := `SELECT ` + allColumnsSQL() + ` FROM products WHERE 1=1`
	params := make(map[string]interface{})

	if filter.Category != "" {
		sql += ` AND category = @category`
		params["category"] = filter.Category
	}

	if filter.Status != "" {
		sql += ` AND status = @status`
		params["status"] = filter.Status
	} else if filter.ActiveOnly {
		sql += ` AND status = @status`
		params["status"] = string(domain.ProductStatusActive)
	}

	// Exclude archived products by default unless specifically filtering for them
	if filter.Status != string(domain.ProductStatusArchived) {
		sql += ` AND status != 'archived'`
	}

	// Pagination using keyset pagination
	if pagination.PageToken != "" {
		sql += ` AND product_id > @page_token`
		params["page_token"] = pagination.PageToken
	}

	sql += ` ORDER BY product_id`

	pageSize := pagination.PageSize
	if pageSize <= 0 {
		pageSize = 20 // default page size
	}
	if pageSize > 100 {
		pageSize = 100 // max page size
	}
	sql += fmt.Sprintf(` LIMIT %d`, pageSize)

	return spanner.Statement{SQL: sql, Params: params}
}

// rowToDTO converts a Spanner row to a ProductDTO.
func (rm *ProductReadModel) rowToDTO(row *spanner.Row, at time.Time) (*contract.ProductDTO, error) {
	var (
		productID            string
		name                 string
		description          string
		category             string
		basePriceNumerator   int64
		basePriceDenominator int64
		discountPercent      spanner.NullNumeric
		discountStartDate    spanner.NullTime
		discountEndDate      spanner.NullTime
		status               string
		createdAt            time.Time
		updatedAt            time.Time
		archivedAt           spanner.NullTime
	)

	if err := row.Columns(
		&productID,
		&name,
		&description,
		&category,
		&basePriceNumerator,
		&basePriceDenominator,
		&discountPercent,
		&discountStartDate,
		&discountEndDate,
		&status,
		&createdAt,
		&updatedAt,
		&archivedAt,
	); err != nil {
		return nil, err
	}

	dto := &contract.ProductDTO{
		ID:                 productID,
		Name:               name,
		Description:        description,
		Category:           category,
		BasePriceNum:       basePriceNumerator,
		BasePriceDenom:     basePriceDenominator,
		Status:             status,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
		EffectivePriceNum:  basePriceNumerator,
		EffectivePriceDenom: basePriceDenominator,
	}

	// Handle discount fields
	if discountPercent.Valid {
		pct, _ := discountPercent.Numeric.Float64()
		dto.DiscountPercent = &pct
	}
	if discountStartDate.Valid {
		dto.DiscountStartDate = &discountStartDate.Time
	}
	if discountEndDate.Valid {
		dto.DiscountEndDate = &discountEndDate.Time
	}

	// Calculate effective price if there's an active discount
	if dto.DiscountPercent != nil && dto.DiscountStartDate != nil && dto.DiscountEndDate != nil {
		if !at.Before(*dto.DiscountStartDate) && at.Before(*dto.DiscountEndDate) {
			dto.HasActiveDiscount = true
			basePrice := domain.NewMoney(basePriceNumerator, basePriceDenominator)
			discountPct := big.NewRat(int64(*dto.DiscountPercent), 1)
			effectivePrice := basePrice.ApplyDiscount(discountPct)
			dto.EffectivePriceNum = effectivePrice.Numerator()
			dto.EffectivePriceDenom = effectivePrice.Denominator()
		}
	}

	return dto, nil
}

// allColumnsSQL returns all column names as a comma-separated SQL string.
func allColumnsSQL() string {
	return `product_id, name, description, category, base_price_numerator, base_price_denominator, 
		discount_percent, discount_start_date, discount_end_date, status, created_at, updated_at, archived_at`
}
