package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCreateProductRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateProductRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: CreateProductRequest{
				Name:                 "Test Product",
				Description:          "A test product",
				Category:             "Electronics",
				BasePriceNumerator:   1999,
				BasePriceDenominator: 100,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			req: CreateProductRequest{
				Name:                 "",
				Description:          "A test product",
				Category:             "Electronics",
				BasePriceNumerator:   1999,
				BasePriceDenominator: 100,
			},
			wantErr: true,
			errMsg:  "invalid product name",
		},
		{
			name: "empty category",
			req: CreateProductRequest{
				Name:                 "Test Product",
				Description:          "A test product",
				Category:             "",
				BasePriceNumerator:   1999,
				BasePriceDenominator: 100,
			},
			wantErr: true,
			errMsg:  "invalid product category",
		},
		{
			name: "zero price numerator",
			req: CreateProductRequest{
				Name:                 "Test Product",
				Description:          "A test product",
				Category:             "Electronics",
				BasePriceNumerator:   0,
				BasePriceDenominator: 100,
			},
			wantErr: true,
			errMsg:  "base price must be positive",
		},
		{
			name: "negative price numerator",
			req: CreateProductRequest{
				Name:                 "Test Product",
				Description:          "A test product",
				Category:             "Electronics",
				BasePriceNumerator:   -100,
				BasePriceDenominator: 100,
			},
			wantErr: true,
			errMsg:  "base price must be positive",
		},
		{
			name: "zero price denominator",
			req: CreateProductRequest{
				Name:                 "Test Product",
				Description:          "A test product",
				Category:             "Electronics",
				BasePriceNumerator:   1999,
				BasePriceDenominator: 0,
			},
			wantErr: true,
			errMsg:  "base price must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateProductRequest(tt.req)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateUpdateProductRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     UpdateProductRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: UpdateProductRequest{
				ProductID:   "123e4567-e89b-12d3-a456-426614174000",
				Name:        "Updated Product",
				Description: "Updated description",
				Category:    "Electronics",
			},
			wantErr: false,
		},
		{
			name: "empty product ID",
			req: UpdateProductRequest{
				ProductID:   "",
				Name:        "Updated Product",
				Description: "Updated description",
				Category:    "Electronics",
			},
			wantErr: true,
			errMsg:  "invalid ID",
		},
		{
			name: "empty name",
			req: UpdateProductRequest{
				ProductID:   "123e4567-e89b-12d3-a456-426614174000",
				Name:        "",
				Description: "Updated description",
				Category:    "Electronics",
			},
			wantErr: true,
			errMsg:  "invalid product name",
		},
		{
			name: "empty category",
			req: UpdateProductRequest{
				ProductID:   "123e4567-e89b-12d3-a456-426614174000",
				Name:        "Updated Product",
				Description: "Updated description",
				Category:    "",
			},
			wantErr: true,
			errMsg:  "invalid product category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUpdateProductRequest(tt.req)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateProductIDRequest(t *testing.T) {
	tests := []struct {
		name      string
		productID string
		wantErr   bool
	}{
		{
			name:      "valid UUID",
			productID: "123e4567-e89b-12d3-a456-426614174000",
			wantErr:   false,
		},
		{
			name:      "valid non-UUID ID",
			productID: "product-123",
			wantErr:   false,
		},
		{
			name:      "empty ID",
			productID: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProductIDRequest(tt.productID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateApplyDiscountRequest(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		req     ApplyDiscountRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request - 10% discount",
			req: ApplyDiscountRequest{
				ProductID:          "123e4567-e89b-12d3-a456-426614174000",
				DiscountPercentage: 10,
				StartDate:          now,
				EndDate:            now.AddDate(0, 1, 0),
			},
			wantErr: false,
		},
		{
			name: "valid request - 50% discount",
			req: ApplyDiscountRequest{
				ProductID:          "123e4567-e89b-12d3-a456-426614174000",
				DiscountPercentage: 50,
				StartDate:          now,
				EndDate:            now.AddDate(0, 0, 7),
			},
			wantErr: false,
		},
		{
			name: "valid request - 100% discount",
			req: ApplyDiscountRequest{
				ProductID:          "123e4567-e89b-12d3-a456-426614174000",
				DiscountPercentage: 100,
				StartDate:          now,
				EndDate:            now.AddDate(0, 0, 1),
			},
			wantErr: false,
		},
		{
			name: "empty product ID",
			req: ApplyDiscountRequest{
				ProductID:          "",
				DiscountPercentage: 10,
				StartDate:          now,
				EndDate:            now.AddDate(0, 1, 0),
			},
			wantErr: true,
			errMsg:  "invalid ID",
		},
		{
			name: "zero discount percentage",
			req: ApplyDiscountRequest{
				ProductID:          "123e4567-e89b-12d3-a456-426614174000",
				DiscountPercentage: 0,
				StartDate:          now,
				EndDate:            now.AddDate(0, 1, 0),
			},
			wantErr: true,
			errMsg:  "discount percentage must be between 0 and 100",
		},
		{
			name: "negative discount percentage",
			req: ApplyDiscountRequest{
				ProductID:          "123e4567-e89b-12d3-a456-426614174000",
				DiscountPercentage: -10,
				StartDate:          now,
				EndDate:            now.AddDate(0, 1, 0),
			},
			wantErr: true,
			errMsg:  "discount percentage must be between 0 and 100",
		},
		{
			name: "discount over 100%",
			req: ApplyDiscountRequest{
				ProductID:          "123e4567-e89b-12d3-a456-426614174000",
				DiscountPercentage: 150,
				StartDate:          now,
				EndDate:            now.AddDate(0, 1, 0),
			},
			wantErr: true,
			errMsg:  "discount percentage must be between 0 and 100",
		},
		{
			name: "end date before start date",
			req: ApplyDiscountRequest{
				ProductID:          "123e4567-e89b-12d3-a456-426614174000",
				DiscountPercentage: 10,
				StartDate:          now.AddDate(0, 1, 0),
				EndDate:            now,
			},
			wantErr: true,
			errMsg:  "discount end date must be after start date",
		},
		{
			name: "end date equals start date",
			req: ApplyDiscountRequest{
				ProductID:          "123e4567-e89b-12d3-a456-426614174000",
				DiscountPercentage: 10,
				StartDate:          now,
				EndDate:            now,
			},
			wantErr: true,
			errMsg:  "discount end date must be after start date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateApplyDiscountRequest(tt.req)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
