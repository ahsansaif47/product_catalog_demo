package contracts

import (
	"product-catalog-service/internal/app/product/domain"

	"cloud.google.com/go/spanner"
)

// ProductRepository defines the interface for product persistence
type ProductRepository interface {
	// InsertMut returns a mutation to insert a product (does not apply)
	InsertMut(product *domain.Product) *spanner.Mutation

	// UpdateMut returns a mutation to update a product (does not apply)
	UpdateMut(product *domain.Product) *spanner.Mutation

	// FindByID retrieves a product by ID
	FindByID(ctx interface{}, productID string) (*domain.Product, error)

	// Exists checks if a product exists
	Exists(ctx interface{}, productID string) (bool, error)
}

// OutboxRepository defines the interface for outbox event persistence
type OutboxRepository interface {
	// InsertMut returns a mutation to insert an outbox event (does not apply)
	InsertMut(event OutboxEvent) *spanner.Mutation
}

// OutboxEvent represents an enriched domain event ready for persistence
type OutboxEvent struct {
	EventID     string
	EventType   string
	AggregateID string
	Payload     map[string]interface{}
}
