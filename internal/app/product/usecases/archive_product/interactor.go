package archive_product

import (
	"context"
	"fmt"

	"github.com/your-username/commitplan"
	"product-catalog-service/internal/app/product/create_product"
	"product-catalog-service/internal/app/product/domain"
)

// ProductReader defines the interface for reading products
type ProductReader interface {
	FindByID(ctx context.Context, productID string) (*domain.Product, error)
}

// ProductWriter defines the interface for writing products
type ProductWriter interface {
	UpdateMut(product *domain.Product) *commitplan.Mutation
}

// EventEnricher enriches domain events for the outbox
type EventEnricher interface {
	EnrichEvent(event domain.DomainEvent) create_product.OutboxEvent
}

// Request represents the archive product request
type Request struct {
	ProductID string
}

// Response represents the archive product response
type Response struct{}

// Interactor handles product archival
type Interactor struct {
	reader    ProductReader
	writer    ProductWriter
	outboxRepo create_product.OutboxRepository
	committer create_product.Committer
	clock     create_product.Clock
	enricher  EventEnricher
}

// NewInteractor creates a new archive product interactor
func NewInteractor(
	reader ProductReader,
	writer ProductWriter,
	outboxRepo create_product.OutboxRepository,
	committer create_product.Committer,
	clock create_product.Clock,
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

// Execute archives a product
func (it *Interactor) Execute(ctx context.Context, req Request) (*Response, error) {
	// Load product
	product, err := it.reader.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	// Archive via domain
	if err := product.Archive(it.clock.Now()); err != nil {
		return nil, err
	}

	// Build commit plan
	plan := commitplan.NewPlan()

	// Add update mutation
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
