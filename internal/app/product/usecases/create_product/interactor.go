package create_product

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/pkg/commitplan"
)

// Clock provides time abstraction
type Clock interface {
	Now() time.Time
}

// Committer applies commit plans
type Committer interface {
	Apply(ctx context.Context, plan *commitplan.Plan) error
}

// ProductRepository defines the repository interface for products
type ProductRepository interface {
	InsertMut(product *domain.Product) *spanner.Mutation
}

// OutboxRepository defines the repository interface for outbox events
type OutboxRepository interface {
	InsertMut(event contracts.OutboxEvent) *spanner.Mutation
}

// Request represents the create product request
type Request struct {
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
}

// Response represents the create product response
type Response struct {
	ProductID string
}

// Interactor handles product creation
type Interactor struct {
	repo      ProductRepository
	outboxRepo OutboxRepository
	committer Committer
	clock     Clock
}

// NewInteractor creates a new create product interactor
func NewInteractor(
	repo ProductRepository,
	outboxRepo OutboxRepository,
	committer Committer,
	clock Clock,
) *Interactor {
	return &Interactor{
		repo:       repo,
		outboxRepo: outboxRepo,
		committer:  committer,
		clock:      clock,
	}
}

// Execute creates a new product
func (it *Interactor) Execute(ctx context.Context, req Request) (*Response, error) {
	// Validate request
	if req.Name == "" {
		return nil, domain.ErrInvalidName
	}
	if req.Category == "" {
		return nil, domain.ErrInvalidCategory
	}

	// Create domain value objects
	basePrice, err := domain.NewMoney(req.BasePriceNumerator, req.BasePriceDenominator)
	if err != nil {
		return nil, err
	}

	// Create product aggregate
	productID := uuid.New().String()
	product, err := domain.NewProduct(
		productID,
		req.Name,
		req.Description,
		req.Category,
		basePrice,
		it.clock.Now(),
	)
	if err != nil {
		return nil, err
	}

	// Build commit plan
	plan := commitplan.NewPlan()

	// Add product mutation
	if mut := it.repo.InsertMut(product); mut != nil {
		plan.Add(mut)
	}

	// Add outbox events for all domain events
	for _, event := range product.DomainEvents() {
		outboxEvent := it.enrichEvent(event)
		outboxMut := it.outboxRepo.InsertMut(outboxEvent)
		if outboxMut != nil {
			plan.Add(outboxMut)
		}
	}

	// Apply the plan
	if err := it.committer.Apply(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to apply commit plan: %w", err)
	}

	return &Response{
		ProductID: productID,
	}, nil
}

// enrichEvent enriches a domain event with metadata
func (it *Interactor) enrichEvent(event domain.DomainEvent) contracts.OutboxEvent {
	payload := make(map[string]interface{})

	// Add common fields
	payload["aggregate_id"] = event.AggregateID()
	payload["event_type"] = event.EventType()
	payload["occurred_at"] = event.OccurredAt().Unix()

	// Add event-specific fields
	switch e := event.(type) {
	case domain.ProductCreatedEvent:
		payload["name"] = e.Name
		payload["category"] = e.Category
		payload["base_price_numerator"] = e.BasePriceNumerator
		payload["base_price_denominator"] = e.BasePriceDenominator
	case domain.ProductUpdatedEvent:
		// No additional fields
	case domain.ProductActivatedEvent:
		// No additional fields
	case domain.ProductDeactivatedEvent:
		// No additional fields
	case domain.ProductArchivedEvent:
		// No additional fields
	case domain.DiscountAppliedEvent:
		payload["discount_percent"] = e.DiscountPercent
		payload["start_date"] = e.StartDate
		payload["end_date"] = e.EndDate
	case domain.DiscountRemovedEvent:
		// No additional fields
	}

	return contracts.OutboxEvent{
		EventID:     uuid.New().String(),
		EventType:   event.EventType(),
		AggregateID: event.AggregateID(),
		Payload:     payload,
	}
}
