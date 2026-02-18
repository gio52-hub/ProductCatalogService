// Package handler implements the gRPC transport layer for the product catalog service.
package handler

import (
	"errors"

	"github.com/product-catalog-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MapDomainErrorToGRPC converts domain errors to gRPC status errors.
func MapDomainErrorToGRPC(err error) error {
	if err == nil {
		return nil
	}

	switch {
	// Not found errors
	case errors.Is(err, domain.ErrProductNotFound):
		return status.Error(codes.NotFound, err.Error())

	// Invalid argument errors
	case errors.Is(err, domain.ErrInvalidID):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrInvalidProductName):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrInvalidProductCategory):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrInvalidBasePrice):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrInvalidDiscountPercentage):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrInvalidDiscountPeriod):
		return status.Error(codes.InvalidArgument, err.Error())

	// Precondition failed errors
	case errors.Is(err, domain.ErrProductNotActive):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrProductArchived):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrProductAlreadyActive):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrProductAlreadyInactive):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrDiscountNotActive):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrDiscountAlreadyExists):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrNoDiscountToRemove):
		return status.Error(codes.FailedPrecondition, err.Error())

	// Default to internal error
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
