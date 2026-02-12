package domain

import "errors"

var (
	// Product errors
	ErrProductNotFound      = errors.New("product not found")
	ErrProductNotActive     = errors.New("product is not active")
	ErrProductAlreadyActive = errors.New("product is already active")
	ErrProductIsArchived    = errors.New("product is archived")

	// Discount errors
	ErrInvalidDiscountPeriod = errors.New("discount period is invalid")
	ErrDiscountOutOfRange    = errors.New("discount must be between 0 and 100")
	ErrNoActiveDiscount      = errors.New("no active discount to remove")

	// Validation errors
	ErrInvalidName        = errors.New("name cannot be empty")
	ErrInvalidCategory    = errors.New("category cannot be empty")
	ErrInvalidPrice       = errors.New("price must be positive")
	ErrInvalidDateRange   = errors.New("end date must be after start date")
	ErrConcurrentModification = errors.New("product was modified by another transaction")
)
