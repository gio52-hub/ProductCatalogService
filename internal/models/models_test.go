package models

import (
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductData_InsertMap(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		data *ProductData
	}{
		{
			name: "complete product data",
			data: &ProductData{
				ProductID:            "product-123",
				Name:                 "Test Product",
				Description:          "A test product description",
				Category:             "Electronics",
				BasePriceNumerator:   1999,
				BasePriceDenominator: 100,
				Status:               "active",
				CreatedAt:            now,
				UpdatedAt:            now,
			},
		},
		{
			name: "product with discount",
			data: &ProductData{
				ProductID:            "product-456",
				Name:                 "Discounted Product",
				Description:          "On sale",
				Category:             "Clothing",
				BasePriceNumerator:   5000,
				BasePriceDenominator: 100,
				DiscountPercent:      spanner.NullNumeric{Valid: true},
				DiscountStartDate:    spanner.NullTime{Time: now, Valid: true},
				DiscountEndDate:      spanner.NullTime{Time: now.AddDate(0, 1, 0), Valid: true},
				Status:               "active",
				CreatedAt:            now,
				UpdatedAt:            now,
			},
		},
		{
			name: "archived product",
			data: &ProductData{
				ProductID:            "product-789",
				Name:                 "Archived Product",
				Description:          "No longer available",
				Category:             "Books",
				BasePriceNumerator:   2500,
				BasePriceDenominator: 100,
				Status:               "archived",
				CreatedAt:            now.AddDate(-1, 0, 0),
				UpdatedAt:            now,
				ArchivedAt:           spanner.NullTime{Time: now, Valid: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.data.InsertMap()

			require.NotNil(t, m)
			assert.Equal(t, tt.data.ProductID, m[ProductID])
			assert.Equal(t, tt.data.Name, m[ProductName])
			assert.Equal(t, tt.data.Description, m[ProductDescription])
			assert.Equal(t, tt.data.Category, m[ProductCategory])
			assert.Equal(t, tt.data.BasePriceNumerator, m[ProductBasePriceNum])
			assert.Equal(t, tt.data.BasePriceDenominator, m[ProductBasePriceDenom])
			assert.Equal(t, tt.data.Status, m[ProductStatus])
			assert.Equal(t, tt.data.CreatedAt, m[ProductCreatedAt])
			assert.Equal(t, tt.data.UpdatedAt, m[ProductUpdatedAt])
		})
	}
}

func TestOutboxEventData_InsertMap(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		data *OutboxEventData
	}{
		{
			name: "pending event",
			data: &OutboxEventData{
				EventID:     "event-123",
				EventType:   "product.created",
				AggregateID: "product-123",
				Status:      StatusPending,
				CreatedAt:   now,
			},
		},
		{
			name: "processed event",
			data: &OutboxEventData{
				EventID:     "event-456",
				EventType:   "product.updated",
				AggregateID: "product-456",
				Status:      StatusProcessed,
				CreatedAt:   now.Add(-time.Hour),
				ProcessedAt: spanner.NullTime{Time: now, Valid: true},
			},
		},
		{
			name: "failed event",
			data: &OutboxEventData{
				EventID:     "event-789",
				EventType:   "product.discount_applied",
				AggregateID: "product-789",
				Status:      StatusFailed,
				CreatedAt:   now.Add(-time.Hour),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.data.InsertMap()

			require.NotNil(t, m)
			assert.Equal(t, tt.data.EventID, m[OutboxEventID])
			assert.Equal(t, tt.data.EventType, m[OutboxEventType])
			assert.Equal(t, tt.data.AggregateID, m[OutboxAggregateID])
			assert.Equal(t, tt.data.Status, m[OutboxStatus])
			assert.Equal(t, tt.data.CreatedAt, m[OutboxCreatedAt])
		})
	}
}

func TestProductAllColumns(t *testing.T) {
	columns := ProductAllColumns()

	expectedColumns := []string{
		ProductID,
		ProductName,
		ProductDescription,
		ProductCategory,
		ProductBasePriceNum,
		ProductBasePriceDenom,
		ProductDiscountPercent,
		ProductDiscountStartDate,
		ProductDiscountEndDate,
		ProductStatus,
		ProductCreatedAt,
		ProductUpdatedAt,
		ProductArchivedAt,
	}

	assert.Equal(t, len(expectedColumns), len(columns))
	for i, col := range expectedColumns {
		assert.Equal(t, col, columns[i])
	}
}

func TestOutboxAllColumns(t *testing.T) {
	columns := OutboxAllColumns()

	expectedColumns := []string{
		OutboxEventID,
		OutboxEventType,
		OutboxAggregateID,
		OutboxPayload,
		OutboxStatus,
		OutboxCreatedAt,
		OutboxProcessedAt,
	}

	assert.Equal(t, len(expectedColumns), len(columns))
	for i, col := range expectedColumns {
		assert.Equal(t, col, columns[i])
	}
}

func TestProductModel_UpdateMut(t *testing.T) {
	model := NewProductModel()

	tests := []struct {
		name      string
		productID string
		updates   map[string]interface{}
		wantNil   bool
	}{
		{
			name:      "empty updates returns nil",
			productID: "product-123",
			updates:   map[string]interface{}{},
			wantNil:   true,
		},
		{
			name:      "nil updates returns nil",
			productID: "product-123",
			updates:   nil,
			wantNil:   true,
		},
		{
			name:      "single field update",
			productID: "product-123",
			updates: map[string]interface{}{
				ProductName: "Updated Name",
			},
			wantNil: false,
		},
		{
			name:      "multiple field update",
			productID: "product-123",
			updates: map[string]interface{}{
				ProductName:        "Updated Name",
				ProductDescription: "Updated description",
				ProductStatus:      "active",
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.UpdateMut(tt.productID, tt.updates)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestOutboxModel_UpdateMut(t *testing.T) {
	model := NewOutboxModel()

	tests := []struct {
		name    string
		eventID string
		updates map[string]interface{}
		wantNil bool
	}{
		{
			name:    "empty updates returns nil",
			eventID: "event-123",
			updates: map[string]interface{}{},
			wantNil: true,
		},
		{
			name:    "single field update",
			eventID: "event-123",
			updates: map[string]interface{}{
				OutboxStatus: StatusProcessed,
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.UpdateMut(tt.eventID, tt.updates)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}
