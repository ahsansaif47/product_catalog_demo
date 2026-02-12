package update_product

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/pkg/commitplan"
)

// ProductReader defines the interface for reading products
type ProductReader interface {
	FindByID(ctx context.Context, productID string) (*domain.Product, error)
}

// ProductWriter defines the interface for writing products
type ProductWriter interface {
	UpdateMut(product *domain.Product) *spanner.Mutation
}

// OutboxRepository defines the repository interface for outbox events
type OutboxRepository interface {
	InsertMut(event contracts.OutboxEvent) *spanner.Mutation
}

// Committer applies commit plans
type Committer interface {
	Apply(ctx context.Context, plan *commitplan.Plan) error
}

// Clock provides time abstraction
type Clock interface {
	Now() time.Time
}

// Request represents the update product request
type Request struct {
	ProductID   string
	Name        string
	Description string
	Category    string
}

// Response represents the update product response
type Response struct{}

// Interactor handles product updates
type Interactor struct {
	reader    ProductReader
	writer    ProductWriter
	outboxRepo OutboxRepository
	committer Committer
	clock     Clock
	enricher  EventEnricher
}

// EventEnricher enriches domain events for the outbox
type EventEnricher interface {
	EnrichEvent(event domain.DomainEvent) contracts.OutboxEvent
}

// NewInteractor creates a new update product interactor
func NewInteractor(
	reader ProductReader,
	writer ProductWriter,
	outboxRepo OutboxRepository,
	committer Committer,
	clock Clock,
	enricher EventEnricher,
) *Interactor {
	return &Interactor{
		reader:    reader,
		writer:    writer,
		outboxRepo: outboxRepo,
		committer:  committer,
		clock:      clock,
		enricher:   enricher,
	}
}

// Execute updates a product
func (it *Interactor) Execute(ctx context.Context, req Request) (*Response, error) {
	// Load product
	product, err := it.reader.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	// Update domain
	if err := product.UpdateDetails(req.Name, req.Description, req.Category, it.clock.Now()); err != nil {
		return nil, err
	}

	// Build commit plan
	plan := commitplan.NewPlan()

	// Add update mutation if there are changes
	if mut := it.writer.UpdateMut(product); mut != nil {
		plan.Add(mut)
	}

	// Add outbox events
	for _, event := range product.DomainEvents() {
		outboxEvent := it.enricher.EnrichEvent(event)
		outboxMut := it.outboxRepo.InsertMut(outboxEvent)
		if outboxMut != nil {
			plan.Add(outboxMut)
		}
	}

	// Apply the plan
	if err := it.committer.Apply(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to apply commit plan: %w", err)
	}

	return &Response{}, nil
}
