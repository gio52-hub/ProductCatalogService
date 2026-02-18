package e2e

import (
	"testing"
	"time"

	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/product-catalog-service/internal/app/product/queries"
	"github.com/product-catalog-service/internal/app/product/usecases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductCreationFlow(t *testing.T) {
	fixture := SetupTestFixture(t)
	ctx := fixture.Context()

	// Test: Create product
	req := usecases.CreateProductRequest{
		Name:                 "Test Product",
		Description:          "A test product description",
		Category:             "Electronics",
		BasePriceNumerator:   1999,
		BasePriceDenominator: 100,
	}

	resp, err := fixture.UseCases.CreateProduct(ctx, req)
	require.NoError(t, err)
	require.NotEmpty(t, resp.ProductID)

	t.Cleanup(func() {
		fixture.CleanupProduct(t, resp.ProductID)
	})

	// Verify: Query returns correct data
	product, err := fixture.Queries.GetProduct(ctx, queries.GetProductRequest{ProductID: resp.ProductID})
	require.NoError(t, err)
	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, "A test product description", product.Description)
	assert.Equal(t, "Electronics", product.Category)
	assert.Equal(t, int64(1999), product.BasePriceNumerator)
	assert.Equal(t, int64(100), product.BasePriceDenominator)
	assert.Equal(t, "draft", product.Status)

	// Verify: Outbox event was created
	events := fixture.GetOutboxEvents(t, resp.ProductID)
	require.Len(t, events, 1)
	assert.Equal(t, "product.created", events[0].EventType)
	assert.Equal(t, "pending", events[0].Status)
}

func TestProductUpdateFlow(t *testing.T) {
	fixture := SetupTestFixture(t)
	ctx := fixture.Context()

	// Setup: Create a product
	createResp, err := fixture.UseCases.CreateProduct(ctx, usecases.CreateProductRequest{
		Name:                 "Original Name",
		Description:          "Original description",
		Category:             "Books",
		BasePriceNumerator:   2500,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		fixture.CleanupProduct(t, createResp.ProductID)
	})

	// Test: Update product
	fixture.AdvanceTime(time.Hour)

	err = fixture.UseCases.UpdateProduct(ctx, usecases.UpdateProductRequest{
		ProductID:   createResp.ProductID,
		Name:        "Updated Name",
		Description: "Updated description",
		Category:    "Fiction",
	})
	require.NoError(t, err)

	// Verify: Query returns updated data
	product, err := fixture.Queries.GetProduct(ctx, queries.GetProductRequest{ProductID: createResp.ProductID})
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", product.Name)
	assert.Equal(t, "Updated description", product.Description)
	assert.Equal(t, "Fiction", product.Category)

	// Verify: Both creation and update events exist
	events := fixture.GetOutboxEvents(t, createResp.ProductID)
	require.Len(t, events, 2)
	assert.Equal(t, "product.created", events[0].EventType)
	assert.Equal(t, "product.updated", events[1].EventType)
}

func TestDiscountApplicationFlow(t *testing.T) {
	fixture := SetupTestFixture(t)
	ctx := fixture.Context()

	// Setup: Create and activate a product
	createResp, err := fixture.UseCases.CreateProduct(ctx, usecases.CreateProductRequest{
		Name:                 "Discounted Product",
		Description:          "A product with discount",
		Category:             "Electronics",
		BasePriceNumerator:   10000,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		fixture.CleanupProduct(t, createResp.ProductID)
	})

	// Activate the product (required for applying discount)
	err = fixture.UseCases.ActivateProduct(ctx, usecases.ActivateProductRequest{ProductID: createResp.ProductID})
	require.NoError(t, err)

	// Test: Apply 20% discount
	now := fixture.Now()
	startDate := now
	endDate := now.Add(7 * 24 * time.Hour)

	err = fixture.UseCases.ApplyDiscount(ctx, usecases.ApplyDiscountRequest{
		ProductID:          createResp.ProductID,
		DiscountPercentage: 20.0,
		StartDate:          startDate,
		EndDate:            endDate,
	})
	require.NoError(t, err)

	// Verify: Effective price is calculated correctly (20% off of $100 = $80)
	product, err := fixture.Queries.GetProduct(ctx, queries.GetProductRequest{ProductID: createResp.ProductID})
	require.NoError(t, err)

	assert.True(t, product.HasActiveDiscount)
	assert.NotNil(t, product.DiscountPercent)
	assert.Equal(t, 20.0, *product.DiscountPercent)

	// Effective price should be 80% of base price
	// Base: 10000/100 = 100, Effective: 8000/100 = 80
	assert.Equal(t, int64(8000), product.EffectivePriceNumerator)
	assert.Equal(t, int64(100), product.EffectivePriceDenominator)

	// Verify: Discount applied event exists
	events := fixture.GetOutboxEvents(t, createResp.ProductID)
	eventTypes := make([]string, len(events))
	for i, e := range events {
		eventTypes[i] = e.EventType
	}
	assert.Contains(t, eventTypes, "product.discount_applied")
}

func TestProductActivationDeactivationFlow(t *testing.T) {
	fixture := SetupTestFixture(t)
	ctx := fixture.Context()

	// Setup: Create a product
	createResp, err := fixture.UseCases.CreateProduct(ctx, usecases.CreateProductRequest{
		Name:                 "Status Test Product",
		Description:          "Testing activation/deactivation",
		Category:             "Test",
		BasePriceNumerator:   500,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		fixture.CleanupProduct(t, createResp.ProductID)
	})

	// Verify initial status is draft
	product, err := fixture.Queries.GetProduct(ctx, queries.GetProductRequest{ProductID: createResp.ProductID})
	require.NoError(t, err)
	assert.Equal(t, "draft", product.Status)

	// Test: Activate product
	fixture.AdvanceTime(time.Minute)
	err = fixture.UseCases.ActivateProduct(ctx, usecases.ActivateProductRequest{ProductID: createResp.ProductID})
	require.NoError(t, err)

	product, err = fixture.Queries.GetProduct(ctx, queries.GetProductRequest{ProductID: createResp.ProductID})
	require.NoError(t, err)
	assert.Equal(t, "active", product.Status)

	// Test: Deactivate product
	fixture.AdvanceTime(time.Minute)
	err = fixture.UseCases.DeactivateProduct(ctx, usecases.DeactivateProductRequest{ProductID: createResp.ProductID})
	require.NoError(t, err)

	product, err = fixture.Queries.GetProduct(ctx, queries.GetProductRequest{ProductID: createResp.ProductID})
	require.NoError(t, err)
	assert.Equal(t, "inactive", product.Status)

	// Verify: Events were created
	events := fixture.GetOutboxEvents(t, createResp.ProductID)
	eventTypes := make([]string, len(events))
	for i, e := range events {
		eventTypes[i] = e.EventType
	}
	assert.Contains(t, eventTypes, "product.activated")
	assert.Contains(t, eventTypes, "product.deactivated")
}

func TestBusinessRuleValidation_CannotApplyDiscountToInactiveProduct(t *testing.T) {
	fixture := SetupTestFixture(t)
	ctx := fixture.Context()

	// Setup: Create a product (in draft status)
	createResp, err := fixture.UseCases.CreateProduct(ctx, usecases.CreateProductRequest{
		Name:                 "Draft Product",
		Description:          "Cannot apply discount",
		Category:             "Test",
		BasePriceNumerator:   1000,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		fixture.CleanupProduct(t, createResp.ProductID)
	})

	// Test: Try to apply discount to draft product (should fail)
	now := fixture.Now()
	err = fixture.UseCases.ApplyDiscount(ctx, usecases.ApplyDiscountRequest{
		ProductID:          createResp.ProductID,
		DiscountPercentage: 10.0,
		StartDate:          now,
		EndDate:            now.Add(24 * time.Hour),
	})

	// Verify: Error is returned
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrProductNotActive)
}

func TestBusinessRuleValidation_CannotActivateArchivedProduct(t *testing.T) {
	fixture := SetupTestFixture(t)
	ctx := fixture.Context()

	// Setup: Create and archive a product
	createResp, err := fixture.UseCases.CreateProduct(ctx, usecases.CreateProductRequest{
		Name:                 "Archived Product",
		Description:          "Will be archived",
		Category:             "Test",
		BasePriceNumerator:   1000,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		fixture.CleanupProduct(t, createResp.ProductID)
	})

	// Archive the product
	err = fixture.UseCases.ArchiveProduct(ctx, usecases.ArchiveProductRequest{ProductID: createResp.ProductID})
	require.NoError(t, err)

	// Test: Try to activate archived product (should fail)
	err = fixture.UseCases.ActivateProduct(ctx, usecases.ActivateProductRequest{ProductID: createResp.ProductID})

	// Verify: Error is returned
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrProductArchived)
}

func TestRemoveDiscountFlow(t *testing.T) {
	fixture := SetupTestFixture(t)
	ctx := fixture.Context()

	// Setup: Create, activate, and add discount
	createResp, err := fixture.UseCases.CreateProduct(ctx, usecases.CreateProductRequest{
		Name:                 "Product With Discount",
		Description:          "Will have discount removed",
		Category:             "Test",
		BasePriceNumerator:   5000,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		fixture.CleanupProduct(t, createResp.ProductID)
	})

	// Activate
	err = fixture.UseCases.ActivateProduct(ctx, usecases.ActivateProductRequest{ProductID: createResp.ProductID})
	require.NoError(t, err)

	// Apply discount
	now := fixture.Now()
	err = fixture.UseCases.ApplyDiscount(ctx, usecases.ApplyDiscountRequest{
		ProductID:          createResp.ProductID,
		DiscountPercentage: 15.0,
		StartDate:          now,
		EndDate:            now.Add(48 * time.Hour),
	})
	require.NoError(t, err)

	// Verify discount is active
	product, err := fixture.Queries.GetProduct(ctx, queries.GetProductRequest{ProductID: createResp.ProductID})
	require.NoError(t, err)
	assert.True(t, product.HasActiveDiscount)

	// Test: Remove discount
	fixture.AdvanceTime(time.Hour)
	err = fixture.UseCases.RemoveDiscount(ctx, usecases.RemoveDiscountRequest{ProductID: createResp.ProductID})
	require.NoError(t, err)

	// Verify: Discount is removed, effective price equals base price
	product, err = fixture.Queries.GetProduct(ctx, queries.GetProductRequest{ProductID: createResp.ProductID})
	require.NoError(t, err)
	assert.False(t, product.HasActiveDiscount)
	assert.Equal(t, product.BasePriceNumerator, product.EffectivePriceNumerator)
	assert.Equal(t, product.BasePriceDenominator, product.EffectivePriceDenominator)

	// Verify: Discount removed event exists
	events := fixture.GetOutboxEvents(t, createResp.ProductID)
	eventTypes := make([]string, len(events))
	for i, e := range events {
		eventTypes[i] = e.EventType
	}
	assert.Contains(t, eventTypes, "product.discount_removed")
}

func TestListProductsWithPagination(t *testing.T) {
	fixture := SetupTestFixture(t)
	ctx := fixture.Context()

	// Setup: Create multiple products
	var productIDs []string
	for i := 0; i < 5; i++ {
		resp, err := fixture.UseCases.CreateProduct(ctx, usecases.CreateProductRequest{
			Name:                 "Paginated Product",
			Description:          "For pagination test",
			Category:             "PaginationTest",
			BasePriceNumerator:   int64(1000 + i*100),
			BasePriceDenominator: 100,
		})
		require.NoError(t, err)
		productIDs = append(productIDs, resp.ProductID)

		// Activate to make them listable
		err = fixture.UseCases.ActivateProduct(ctx, usecases.ActivateProductRequest{ProductID: resp.ProductID})
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		for _, id := range productIDs {
			fixture.CleanupProduct(t, id)
		}
	})

	// Test: List with page size of 2
	result, err := fixture.Queries.ListProducts(ctx, queries.ListProductsRequest{
		Category:   "PaginationTest",
		ActiveOnly: true,
		PageSize:   2,
	})
	require.NoError(t, err)
	assert.Len(t, result.Products, 2)
	assert.NotEmpty(t, result.NextPageToken)

	// Test: Get next page
	result2, err := fixture.Queries.ListProducts(ctx, queries.ListProductsRequest{
		Category:   "PaginationTest",
		ActiveOnly: true,
		PageSize:   2,
		PageToken:  result.NextPageToken,
	})
	require.NoError(t, err)
	assert.Len(t, result2.Products, 2)

	// Ensure different products
	assert.NotEqual(t, result.Products[0].ID, result2.Products[0].ID)
}

func TestOutboxEventCreation(t *testing.T) {
	fixture := SetupTestFixture(t)
	ctx := fixture.Context()

	// Create a product
	createResp, err := fixture.UseCases.CreateProduct(ctx, usecases.CreateProductRequest{
		Name:                 "Outbox Test Product",
		Description:          "Testing outbox events",
		Category:             "Test",
		BasePriceNumerator:   2000,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		fixture.CleanupProduct(t, createResp.ProductID)
	})

	// Verify: Outbox event was created
	events := fixture.GetOutboxEvents(t, createResp.ProductID)
	require.Len(t, events, 1)

	event := events[0]
	assert.NotEmpty(t, event.EventID)
	assert.Equal(t, "product.created", event.EventType)
	assert.Equal(t, createResp.ProductID, event.AggregateID)
	assert.Equal(t, "pending", event.Status)
	assert.False(t, event.CreatedAt.IsZero())
}
