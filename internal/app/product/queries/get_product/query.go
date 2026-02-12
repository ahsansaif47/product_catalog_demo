package get_product

import (
	"context"
	"product-catalog-service/internal/app/product/contracts"
)

// ReadModel defines the interface for reading products
type ReadModel interface {
	GetProduct(ctx context.Context, productID string) (*contracts.ProductDTO, error)
}

// Request represents the get product query request
type Request struct {
	ProductID string
}

// Response represents the get product query response
type Response struct {
	Product *contracts.ProductDTO
}

// Query handles getting a product by ID
type Query struct {
	readModel ReadModel
}

// NewQuery creates a new get product query
func NewQuery(readModel ReadModel) *Query {
	return &Query{
		readModel: readModel,
	}
}

// Execute retrieves a product by ID
func (q *Query) Execute(ctx context.Context, req Request) (*Response, error) {
	product, err := q.readModel.GetProduct(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	return &Response{
		Product: product,
	}, nil
}
