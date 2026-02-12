package services

import (
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/queries/get_product"
	"product-catalog-service/internal/app/product/queries/list_products"
	"product-catalog-service/internal/app/product/repo"
	"product-catalog-service/internal/app/product/usecases/activate_product"
	"product-catalog-service/internal/app/product/usecases/apply_discount"
	"product-catalog-service/internal/app/product/usecases/archive_product"
	"product-catalog-service/internal/app/product/usecases/create_product"
	"product-catalog-service/internal/app/product/usecases/deactivate_product"
	"product-catalog-service/internal/app/product/usecases/remove_discount"
	"product-catalog-service/internal/app/product/usecases/update_product"
	"product-catalog-service/internal/pkg/clock"
	"product-catalog-service/internal/pkg/committer"
	"product-catalog-service/internal/transport/grpc/product"
)

// Container holds all the service dependencies
type Container struct {
	// Infrastructure
	Clock     clock.Clock
	Committer *committer.Committer

	// Repositories
	ProductRepo     *repo.ProductRepo
	OutboxRepo      *repo.OutboxRepo
	ProductReadModel *repo.ProductReadModel

	// Event Enricher
	EventEnricher *EventEnricher

	// Usecases
	CreateProductInteractor      *create_product.Interactor
	UpdateProductInteractor      *update_product.Interactor
	ActivateProductInteractor    *activate_product.Interactor
	DeactivateProductInteractor  *deactivate_product.Interactor
	ApplyDiscountInteractor      *apply_discount.Interactor
	RemoveDiscountInteractor     *remove_discount.Interactor
	ArchiveProductInteractor     *archive_product.Interactor

	// Queries
	GetProductQuery   *get_product.Query
	ListProductsQuery *list_products.Query

	// Handlers
	ProductHandlers *product.Handlers
}

// NewContainer creates a new dependency injection container
func NewContainer(spannerClient *spanner.Client) *Container {
	// Infrastructure
	clk := clock.NewRealClock()
	committer := committer.NewCommitter(spannerClient)

	// Repositories
	productRepo := repo.NewProductRepo(spannerClient)
	outboxRepo := repo.NewOutboxRepo(spannerClient)
	productReadModel := repo.NewProductReadModel(spannerClient)

	// Event Enricher
	eventEnricher := NewEventEnricher()

	// Usecases
	createProductInteractor := create_product.NewInteractor(
		productRepo,
		outboxRepo,
		committer,
		clk,
	)

	updateProductInteractor := update_product.NewInteractor(
		productRepo,
		productRepo,
		outboxRepo,
		committer,
		clk,
		eventEnricher,
	)

	activateProductInteractor := activate_product.NewInteractor(
		productRepo,
		productRepo,
		outboxRepo,
		committer,
		clk,
		eventEnricher,
	)

	deactivateProductInteractor := deactivate_product.NewInteractor(
		productRepo,
		productRepo,
		outboxRepo,
		committer,
		clk,
		eventEnricher,
	)

	applyDiscountInteractor := apply_discount.NewInteractor(
		productRepo,
		productRepo,
		outboxRepo,
		committer,
		clk,
		eventEnricher,
	)

	removeDiscountInteractor := remove_discount.NewInteractor(
		productRepo,
		productRepo,
		outboxRepo,
		committer,
		clk,
		eventEnricher,
	)

	archiveProductInteractor := archive_product.NewInteractor(
		productRepo,
		productRepo,
		outboxRepo,
		committer,
		clk,
		eventEnricher,
	)

	// Queries
	getProductQuery := get_product.NewQuery(productReadModel)
	listProductsQuery := list_products.NewQuery(productReadModel)

	// Handlers
	productHandlers := product.NewHandlers(
		createProductInteractor,
		updateProductInteractor,
		activateProductInteractor,
		deactivateProductInteractor,
		applyDiscountInteractor,
		removeDiscountInteractor,
		archiveProductInteractor,
		getProductQuery,
		listProductsQuery,
	)

	return &Container{
		Clock:                    clk,
		Committer:                committer,
		ProductRepo:              productRepo,
		OutboxRepo:               outboxRepo,
		ProductReadModel:          productReadModel,
		EventEnricher:            eventEnricher,
		CreateProductInteractor:    createProductInteractor,
		UpdateProductInteractor:    updateProductInteractor,
		ActivateProductInteractor:  activateProductInteractor,
		DeactivateProductInteractor: deactivateProductInteractor,
		ApplyDiscountInteractor:   applyDiscountInteractor,
		RemoveDiscountInteractor:  removeDiscountInteractor,
		ArchiveProductInteractor:   archiveProductInteractor,
		GetProductQuery:           getProductQuery,
		ListProductsQuery:         listProductsQuery,
		ProductHandlers:          productHandlers,
	}
}

// EventEnricher enriches domain events for the outbox
type EventEnricher struct {
	clock clock.Clock
}

// NewEventEnricher creates a new event enricher
func NewEventEnricher() *EventEnricher {
	return &EventEnricher{
		clock: clock.NewRealClock(),
	}
}

// EnrichEvent enriches a domain event with metadata
func (e *EventEnricher) EnrichEvent(domainEvent domain.DomainEvent) contracts.OutboxEvent {
	// Extract domain event information
	var (
		aggregateID string
		eventType   string
		occurredAt  time.Time
	)

	switch ev := domainEvent.(type) {
	case domain.ProductCreatedEvent:
		aggregateID = ev.AggregateID()
		eventType = ev.EventType()
		occurredAt = ev.OccurredAt()
	case domain.ProductUpdatedEvent:
		aggregateID = ev.AggregateID()
		eventType = ev.EventType()
		occurredAt = ev.OccurredAt()
	case domain.ProductActivatedEvent:
		aggregateID = ev.AggregateID()
		eventType = ev.EventType()
		occurredAt = ev.OccurredAt()
	case domain.ProductDeactivatedEvent:
		aggregateID = ev.AggregateID()
		eventType = ev.EventType()
		occurredAt = ev.OccurredAt()
	case domain.ProductArchivedEvent:
		aggregateID = ev.AggregateID()
		eventType = ev.EventType()
		occurredAt = ev.OccurredAt()
	case domain.DiscountAppliedEvent:
		aggregateID = ev.AggregateID()
		eventType = ev.EventType()
		occurredAt = ev.OccurredAt()
	case domain.DiscountRemovedEvent:
		aggregateID = ev.AggregateID()
		eventType = ev.EventType()
		occurredAt = ev.OccurredAt()
	default:
		return contracts.OutboxEvent{}
	}

	payload := map[string]interface{}{
		"aggregate_id": aggregateID,
		"event_type":   eventType,
		"occurred_at":  occurredAt.Unix(),
	}

	return contracts.OutboxEvent{
		EventID:     uuid.New().String(),
		EventType:   eventType,
		AggregateID: aggregateID,
		Payload:     payload,
	}
}
