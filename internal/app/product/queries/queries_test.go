package queries

import (
	"testing"
	"time"

	"github.com/product-catalog-service/internal/app/product/contracts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductResponseFromDTO(t *testing.T) {
	tests := []struct {
		name     string
		dto      *contracts.ProductDTO
		wantNil  bool
		expected *ProductResponse
	}{
		{
			name:    "nil dto returns nil",
			dto:     nil,
			wantNil: true,
		},
		{
			name: "valid dto without discount",
			dto: &contracts.ProductDTO{
				ID:                  "product-123",
				Name:                "Test Product",
				Description:         "A description",
				Category:            "Electronics",
				BasePriceNum:        1999,
				BasePriceDenom:      100,
				EffectivePriceNum:   1999,
				EffectivePriceDenom: 100,
				HasActiveDiscount:   false,
				Status:              "active",
				CreatedAt:           time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:           time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			wantNil: false,
			expected: &ProductResponse{
				ID:                        "product-123",
				Name:                      "Test Product",
				Description:               "A description",
				Category:                  "Electronics",
				BasePriceNumerator:        1999,
				BasePriceDenominator:      100,
				EffectivePriceNumerator:   1999,
				EffectivePriceDenominator: 100,
				HasActiveDiscount:         false,
				Status:                    "active",
				CreatedAt:                 time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:                 time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "valid dto with active discount",
			dto: &contracts.ProductDTO{
				ID:                  "product-456",
				Name:                "Discounted Product",
				Description:         "On sale",
				Category:            "Clothing",
				BasePriceNum:        5000,
				BasePriceDenom:      100,
				EffectivePriceNum:   4000,
				EffectivePriceDenom: 100,
				DiscountPercent:     ptrFloat64(20.0),
				DiscountStartDate:   ptrTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				DiscountEndDate:     ptrTime(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)),
				HasActiveDiscount:   true,
				Status:              "active",
				CreatedAt:           time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:           time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			},
			wantNil: false,
			expected: &ProductResponse{
				ID:                        "product-456",
				Name:                      "Discounted Product",
				Description:               "On sale",
				Category:                  "Clothing",
				BasePriceNumerator:        5000,
				BasePriceDenominator:      100,
				EffectivePriceNumerator:   4000,
				EffectivePriceDenominator: 100,
				DiscountPercent:           ptrFloat64(20.0),
				DiscountStartDate:         ptrTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				DiscountEndDate:           ptrTime(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)),
				HasActiveDiscount:         true,
				Status:                    "active",
				CreatedAt:                 time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:                 time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := productResponseFromDTO(tt.dto)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.Name, result.Name)
				assert.Equal(t, tt.expected.Description, result.Description)
				assert.Equal(t, tt.expected.Category, result.Category)
				assert.Equal(t, tt.expected.BasePriceNumerator, result.BasePriceNumerator)
				assert.Equal(t, tt.expected.BasePriceDenominator, result.BasePriceDenominator)
				assert.Equal(t, tt.expected.EffectivePriceNumerator, result.EffectivePriceNumerator)
				assert.Equal(t, tt.expected.EffectivePriceDenominator, result.EffectivePriceDenominator)
				assert.Equal(t, tt.expected.HasActiveDiscount, result.HasActiveDiscount)
				assert.Equal(t, tt.expected.Status, result.Status)
			}
		})
	}
}

func TestListProductsResponseFromDTOs(t *testing.T) {
	tests := []struct {
		name           string
		result         *contracts.ListProductsResult
		expectedCount  int
		expectedTotal  int64
		expectedToken  string
	}{
		{
			name:           "nil result returns empty",
			result:         nil,
			expectedCount:  0,
			expectedTotal:  0,
			expectedToken:  "",
		},
		{
			name: "empty products list",
			result: &contracts.ListProductsResult{
				Products:      []*contracts.ProductDTO{},
				NextPageToken: "",
				TotalCount:    0,
			},
			expectedCount: 0,
			expectedTotal: 0,
			expectedToken: "",
		},
		{
			name: "single product",
			result: &contracts.ListProductsResult{
				Products: []*contracts.ProductDTO{
					{
						ID:                  "product-1",
						Name:                "Product One",
						Category:            "Electronics",
						BasePriceNum:        1000,
						BasePriceDenom:      100,
						EffectivePriceNum:   1000,
						EffectivePriceDenom: 100,
						Status:              "active",
						CreatedAt:           time.Now(),
					},
				},
				NextPageToken: "",
				TotalCount:    1,
			},
			expectedCount: 1,
			expectedTotal: 1,
			expectedToken: "",
		},
		{
			name: "multiple products with pagination",
			result: &contracts.ListProductsResult{
				Products: []*contracts.ProductDTO{
					{ID: "product-1", Name: "Product One", Category: "A", BasePriceNum: 100, BasePriceDenom: 1, EffectivePriceNum: 100, EffectivePriceDenom: 1, Status: "active", CreatedAt: time.Now()},
					{ID: "product-2", Name: "Product Two", Category: "B", BasePriceNum: 200, BasePriceDenom: 1, EffectivePriceNum: 200, EffectivePriceDenom: 1, Status: "active", CreatedAt: time.Now()},
					{ID: "product-3", Name: "Product Three", Category: "A", BasePriceNum: 300, BasePriceDenom: 1, EffectivePriceNum: 300, EffectivePriceDenom: 1, Status: "draft", CreatedAt: time.Now()},
				},
				NextPageToken: "next-page-token-123",
				TotalCount:    50,
			},
			expectedCount: 3,
			expectedTotal: 50,
			expectedToken: "next-page-token-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := listProductsResponseFromDTOs(tt.result)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedCount, len(result.Products))
			assert.Equal(t, tt.expectedTotal, result.TotalCount)
			assert.Equal(t, tt.expectedToken, result.NextPageToken)
		})
	}
}

func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrTime(v time.Time) *time.Time {
	return &v
}
