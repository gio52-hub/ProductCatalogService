package handler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/product-catalog-service/proto/product/v1"
)

func TestValidateCreateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *pb.CreateProductRequest
		wantErr error
	}{
		{
			name: "valid request",
			req: &pb.CreateProductRequest{
				Name:        "Test Product",
				Description: "A test product",
				Category:    "Electronics",
				BasePrice:   &pb.Money{Numerator: 1999, Denominator: 100},
			},
			wantErr: nil,
		},
		{
			name: "valid request with minimal fields",
			req: &pb.CreateProductRequest{
				Name:      "Product",
				Category:  "Category",
				BasePrice: &pb.Money{Numerator: 100, Denominator: 1},
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			req: &pb.CreateProductRequest{
				Name:      "",
				Category:  "Electronics",
				BasePrice: &pb.Money{Numerator: 1999, Denominator: 100},
			},
			wantErr: ErrNameRequired,
		},
		{
			name: "empty category",
			req: &pb.CreateProductRequest{
				Name:      "Test Product",
				Category:  "",
				BasePrice: &pb.Money{Numerator: 1999, Denominator: 100},
			},
			wantErr: ErrCategoryRequired,
		},
		{
			name: "nil base price",
			req: &pb.CreateProductRequest{
				Name:      "Test Product",
				Category:  "Electronics",
				BasePrice: nil,
			},
			wantErr: ErrBasePriceRequired,
		},
		{
			name: "zero numerator",
			req: &pb.CreateProductRequest{
				Name:      "Test Product",
				Category:  "Electronics",
				BasePrice: &pb.Money{Numerator: 0, Denominator: 100},
			},
			wantErr: ErrInvalidBasePrice,
		},
		{
			name: "negative numerator",
			req: &pb.CreateProductRequest{
				Name:      "Test Product",
				Category:  "Electronics",
				BasePrice: &pb.Money{Numerator: -100, Denominator: 100},
			},
			wantErr: ErrInvalidBasePrice,
		},
		{
			name: "zero denominator",
			req: &pb.CreateProductRequest{
				Name:      "Test Product",
				Category:  "Electronics",
				BasePrice: &pb.Money{Numerator: 100, Denominator: 0},
			},
			wantErr: ErrInvalidBasePrice,
		},
		{
			name: "negative denominator",
			req: &pb.CreateProductRequest{
				Name:      "Test Product",
				Category:  "Electronics",
				BasePrice: &pb.Money{Numerator: 100, Denominator: -1},
			},
			wantErr: ErrInvalidBasePrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateRequest(tt.req)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateUpdateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *pb.UpdateProductRequest
		wantErr error
	}{
		{
			name: "valid request",
			req: &pb.UpdateProductRequest{
				ProductId:   "product-123",
				Name:        "Updated Product",
				Description: "Updated description",
				Category:    "Electronics",
			},
			wantErr: nil,
		},
		{
			name: "empty product ID",
			req: &pb.UpdateProductRequest{
				ProductId:   "",
				Name:        "Updated Product",
				Description: "Updated description",
				Category:    "Electronics",
			},
			wantErr: ErrProductIDRequired,
		},
		{
			name: "empty name",
			req: &pb.UpdateProductRequest{
				ProductId:   "product-123",
				Name:        "",
				Description: "Updated description",
				Category:    "Electronics",
			},
			wantErr: ErrNameRequired,
		},
		{
			name: "empty category",
			req: &pb.UpdateProductRequest{
				ProductId:   "product-123",
				Name:        "Updated Product",
				Description: "Updated description",
				Category:    "",
			},
			wantErr: ErrCategoryRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUpdateRequest(tt.req)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateApplyDiscountRequest(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	tests := []struct {
		name    string
		req     *pb.ApplyDiscountRequest
		wantErr error
	}{
		{
			name: "valid request - 10% discount",
			req: &pb.ApplyDiscountRequest{
				ProductId:          "product-123",
				DiscountPercentage: 10,
				StartDate:          timestamppb.New(now),
				EndDate:            timestamppb.New(future),
			},
			wantErr: nil,
		},
		{
			name: "valid request - 50% discount",
			req: &pb.ApplyDiscountRequest{
				ProductId:          "product-123",
				DiscountPercentage: 50,
				StartDate:          timestamppb.New(now),
				EndDate:            timestamppb.New(future),
			},
			wantErr: nil,
		},
		{
			name: "valid request - 100% discount",
			req: &pb.ApplyDiscountRequest{
				ProductId:          "product-123",
				DiscountPercentage: 100,
				StartDate:          timestamppb.New(now),
				EndDate:            timestamppb.New(future),
			},
			wantErr: nil,
		},
		{
			name: "empty product ID",
			req: &pb.ApplyDiscountRequest{
				ProductId:          "",
				DiscountPercentage: 10,
				StartDate:          timestamppb.New(now),
				EndDate:            timestamppb.New(future),
			},
			wantErr: ErrProductIDRequired,
		},
		{
			name: "zero discount percentage",
			req: &pb.ApplyDiscountRequest{
				ProductId:          "product-123",
				DiscountPercentage: 0,
				StartDate:          timestamppb.New(now),
				EndDate:            timestamppb.New(future),
			},
			wantErr: ErrInvalidDiscount,
		},
		{
			name: "negative discount percentage",
			req: &pb.ApplyDiscountRequest{
				ProductId:          "product-123",
				DiscountPercentage: -10,
				StartDate:          timestamppb.New(now),
				EndDate:            timestamppb.New(future),
			},
			wantErr: ErrInvalidDiscount,
		},
		{
			name: "discount over 100%",
			req: &pb.ApplyDiscountRequest{
				ProductId:          "product-123",
				DiscountPercentage: 150,
				StartDate:          timestamppb.New(now),
				EndDate:            timestamppb.New(future),
			},
			wantErr: ErrInvalidDiscount,
		},
		{
			name: "nil start date",
			req: &pb.ApplyDiscountRequest{
				ProductId:          "product-123",
				DiscountPercentage: 10,
				StartDate:          nil,
				EndDate:            timestamppb.New(future),
			},
			wantErr: ErrStartDateRequired,
		},
		{
			name: "nil end date",
			req: &pb.ApplyDiscountRequest{
				ProductId:          "product-123",
				DiscountPercentage: 10,
				StartDate:          timestamppb.New(now),
				EndDate:            nil,
			},
			wantErr: ErrEndDateRequired,
		},
		{
			name: "end date before start date",
			req: &pb.ApplyDiscountRequest{
				ProductId:          "product-123",
				DiscountPercentage: 10,
				StartDate:          timestamppb.New(now),
				EndDate:            timestamppb.New(past),
			},
			wantErr: ErrEndDateBeforeStartDate,
		},
		{
			name: "end date equals start date",
			req: &pb.ApplyDiscountRequest{
				ProductId:          "product-123",
				DiscountPercentage: 10,
				StartDate:          timestamppb.New(now),
				EndDate:            timestamppb.New(now),
			},
			wantErr: ErrEndDateBeforeStartDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateApplyDiscountRequest(tt.req)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
