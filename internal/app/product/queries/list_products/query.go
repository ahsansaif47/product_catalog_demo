package list_products

import (
	"context"
	"product-catalog-service/internal/app/product/contracts"
)

// ReadModel defines the interface for reading products
type ReadModel interface {
	ListProducts(ctx context.Context, filter contracts.ListProductsFilter) (*contracts.PaginatedProductsDTO, error)
}

// Request represents the list products query request
type Request struct {
	Category  string
	PageSize  int
	PageToken string
	Status    string
}

// Response represents the list products query response
type Response struct {
	Products      []*contracts.ProductDTO
	NextPageToken string
}

// Query handles listing products
type Query struct {
	readModel ReadModel
}

// NewQuery creates a new list products query
func NewQuery(readModel ReadModel) *Query {
	return &Query{
		readModel: readModel,
	}
}

// Execute retrieves a list of products
func (q *Query) Execute(ctx context.Context, req Request) (*Response, error) {
	filter := contracts.ListProductsFilter{
		Category:  req.Category,
		PageSize:  req.PageSize,
		PageToken: req.PageToken,
		Status:    req.Status,
	}

	result, err := q.readModel.ListProducts(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &Response{
		Products:      result.Products,
		NextPageToken: result.NextPageToken,
	}, nil
}
