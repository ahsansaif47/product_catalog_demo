package contracts

import (
	"context"
)

// ProductReadModel defines the interface for product queries
type ProductReadModel interface {
	// GetProduct retrieves a product by ID with effective price
	GetProduct(ctx context.Context, productID string) (*ProductDTO, error)

	// ListProducts retrieves a paginated list of products
	ListProducts(ctx context.Context, filter ListProductsFilter) (*PaginatedProductsDTO, error)
}

// ProductDTO represents a product in the read model
type ProductDTO struct {
	ProductID          string
	Name               string
	Description        string
	Category           string
	BasePriceNumerator int64
	BasePriceDenominator int64

	// Effective price after discount
	EffectivePriceNumerator   int64
	EffectivePriceDenominator int64

	// Discount information (if active)
	HasDiscount       bool
	DiscountPercent   *int64
	DiscountStartDate *int64
	DiscountEndDate   *int64

	Status        string
	CreatedAtSec  int64
	UpdatedAtSec  int64
}

// PaginatedProductsDTO represents a paginated list of products
type PaginatedProductsDTO struct {
	Products     []*ProductDTO
	NextPageToken string
}

// ListProductsFilter represents filters for listing products
type ListProductsFilter struct {
	Category  string // Optional filter by category
	PageSize  int
	PageToken string
	Status    string // Optional filter by status
}
