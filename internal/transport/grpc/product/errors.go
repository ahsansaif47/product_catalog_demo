package product

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"product-catalog-service/internal/app/product/domain"
)

// mapDomainErrorToGRPC converts domain errors to gRPC status errors
func mapDomainErrorToGRPC(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		return status.Error(codes.NotFound, "product not found")
	case errors.Is(err, domain.ErrProductNotActive):
		return status.Error(codes.FailedPrecondition, "product is not active")
	case errors.Is(err, domain.ErrProductAlreadyActive):
		return status.Error(codes.FailedPrecondition, "product is already active")
	case errors.Is(err, domain.ErrProductIsArchived):
		return status.Error(codes.FailedPrecondition, "product is archived")
	case errors.Is(err, domain.ErrInvalidDiscountPeriod):
		return status.Error(codes.InvalidArgument, "invalid discount period")
	case errors.Is(err, domain.ErrDiscountOutOfRange):
		return status.Error(codes.OutOfRange, "discount must be between 0 and 100")
	case errors.Is(err, domain.ErrNoActiveDiscount):
		return status.Error(codes.FailedPrecondition, "no active discount to remove")
	case errors.Is(err, domain.ErrInvalidName):
		return status.Error(codes.InvalidArgument, "name cannot be empty")
	case errors.Is(err, domain.ErrInvalidCategory):
		return status.Error(codes.InvalidArgument, "category cannot be empty")
	case errors.Is(err, domain.ErrInvalidPrice):
		return status.Error(codes.InvalidArgument, "price must be positive")
	case errors.Is(err, domain.ErrInvalidDateRange):
		return status.Error(codes.InvalidArgument, "end date must be after start date")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

// Validation functions for proto requests

func validateCreateProductRequest(req interface{}) error {
	// Type assertion would be done here in actual implementation
	// For now, we'll validate directly in the handler
	return nil
}

func validateUpdateProductRequest(req interface{}) error {
	return nil
}

func validateApplyDiscountRequest(req interface{}) error {
	return nil
}
