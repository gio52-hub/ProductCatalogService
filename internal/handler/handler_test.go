package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/product-catalog-service/internal/domain"
	pb "github.com/product-catalog-service/proto/product/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapDomainErrorToGRPC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		inputError   error
		expectedCode codes.Code
	}{
		{
			name:         "product not found",
			inputError:   domain.ErrProductNotFound,
			expectedCode: codes.NotFound,
		},
		{
			name:         "product already active",
			inputError:   domain.ErrProductAlreadyActive,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name:         "product already inactive",
			inputError:   domain.ErrProductAlreadyInactive,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name:         "product archived",
			inputError:   domain.ErrProductArchived,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name:         "invalid product name",
			inputError:   domain.ErrInvalidProductName,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "invalid product category",
			inputError:   domain.ErrInvalidProductCategory,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "invalid base price",
			inputError:   domain.ErrInvalidBasePrice,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "invalid discount percentage",
			inputError:   domain.ErrInvalidDiscountPercentage,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "invalid discount period",
			inputError:   domain.ErrInvalidDiscountPeriod,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "no discount to remove",
			inputError:   domain.ErrNoDiscountToRemove,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name:         "generic error",
			inputError:   errors.New("some internal error"),
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			grpcErr := MapDomainErrorToGRPC(tt.inputError)
			st, ok := status.FromError(grpcErr)
			assert.True(t, ok)
			assert.Equal(t, tt.expectedCode, st.Code())
		})
	}
}

func TestHandler_CreateProduct_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		request     *pb.CreateProductRequest
		expectError bool
		errorCode   codes.Code
	}{
		{
			name: "missing name",
			request: &pb.CreateProductRequest{
				Name:        "",
				Description: "Test description",
				Category:    "Test Category",
				BasePrice: &pb.Money{
					Numerator:   1000,
					Denominator: 100,
				},
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name: "missing category",
			request: &pb.CreateProductRequest{
				Name:        "Test Product",
				Description: "Test description",
				Category:    "",
				BasePrice: &pb.Money{
					Numerator:   1000,
					Denominator: 100,
				},
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name: "missing base price",
			request: &pb.CreateProductRequest{
				Name:        "Test Product",
				Description: "Test description",
				Category:    "Test Category",
				BasePrice:   nil,
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name: "zero denominator",
			request: &pb.CreateProductRequest{
				Name:        "Test Product",
				Description: "Test description",
				Category:    "Test Category",
				BasePrice: &pb.Money{
					Numerator:   1000,
					Denominator: 0,
				},
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name: "negative price",
			request: &pb.CreateProductRequest{
				Name:        "Test Product",
				Description: "Test description",
				Category:    "Test Category",
				BasePrice: &pb.Money{
					Numerator:   -1000,
					Denominator: 100,
				},
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
	}

	handler := NewHandler(nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := handler.CreateProduct(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errorCode, st.Code())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_ActivateProduct_Validation(t *testing.T) {
	t.Parallel()

	handler := NewHandler(nil, nil)

	_, err := handler.ActivateProduct(context.Background(), &pb.ActivateProductRequest{
		ProductId: "",
	})

	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestHandler_DeactivateProduct_Validation(t *testing.T) {
	t.Parallel()

	handler := NewHandler(nil, nil)

	_, err := handler.DeactivateProduct(context.Background(), &pb.DeactivateProductRequest{
		ProductId: "",
	})

	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestHandler_ArchiveProduct_Validation(t *testing.T) {
	t.Parallel()

	handler := NewHandler(nil, nil)

	_, err := handler.ArchiveProduct(context.Background(), &pb.ArchiveProductRequest{
		ProductId: "",
	})

	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestHandler_RemoveDiscount_Validation(t *testing.T) {
	t.Parallel()

	handler := NewHandler(nil, nil)

	_, err := handler.RemoveDiscount(context.Background(), &pb.RemoveDiscountRequest{
		ProductId: "",
	})

	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestHandler_GetProduct_Validation(t *testing.T) {
	t.Parallel()

	handler := NewHandler(nil, nil)

	_, err := handler.GetProduct(context.Background(), &pb.GetProductRequest{
		ProductId: "",
	})

	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestHandler_UpdateProduct_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		request     *pb.UpdateProductRequest
		expectError bool
		errorCode   codes.Code
	}{
		{
			name: "missing product_id",
			request: &pb.UpdateProductRequest{
				ProductId:   "",
				Name:        "Updated Name",
				Description: "Updated description",
				Category:    "Updated Category",
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
	}

	handler := NewHandler(nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := handler.UpdateProduct(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errorCode, st.Code())
			}
		})
	}
}

func TestHandler_ApplyDiscount_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		request     *pb.ApplyDiscountRequest
		expectError bool
		errorCode   codes.Code
	}{
		{
			name: "missing product_id",
			request: &pb.ApplyDiscountRequest{
				ProductId:          "",
				DiscountPercentage: 10.0,
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name: "invalid discount percentage - zero",
			request: &pb.ApplyDiscountRequest{
				ProductId:          "test-id",
				DiscountPercentage: 0,
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name: "invalid discount percentage - over 100",
			request: &pb.ApplyDiscountRequest{
				ProductId:          "test-id",
				DiscountPercentage: 150,
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name: "invalid discount percentage - negative",
			request: &pb.ApplyDiscountRequest{
				ProductId:          "test-id",
				DiscountPercentage: -10,
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
	}

	handler := NewHandler(nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := handler.ApplyDiscount(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errorCode, st.Code())
			}
		})
	}
}
