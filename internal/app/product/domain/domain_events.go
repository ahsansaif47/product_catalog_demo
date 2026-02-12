package domain

import "time"

// DomainEvent represents a domain event
type DomainEvent interface {
	AggregateID() string
	EventType() string
	OccurredAt() time.Time
}

// BaseEvent provides common event fields
type BaseEvent struct {
	aggregateID string
	eventType   string
	occurredAt  time.Time
}

func NewBaseEvent(aggregateID, eventType string) BaseEvent {
	return BaseEvent{
		aggregateID: aggregateID,
		eventType:   eventType,
		occurredAt:  time.Now(),
	}
}

func (e BaseEvent) AggregateID() string { return e.aggregateID }
func (e BaseEvent) EventType() string   { return e.eventType }
func (e BaseEvent) OccurredAt() time.Time { return e.occurredAt }

// ProductCreatedEvent is emitted when a new product is created
type ProductCreatedEvent struct {
	BaseEvent
	Name        string
	Category    string
	BasePriceNumerator   int64
	BasePriceDenominator int64
}

func NewProductCreatedEvent(aggregateID, name, category string, numerator, denominator int64) ProductCreatedEvent {
	return ProductCreatedEvent{
		BaseEvent:            NewBaseEvent(aggregateID, "product.created"),
		Name:                 name,
		Category:             category,
		BasePriceNumerator:   numerator,
		BasePriceDenominator: denominator,
	}
}

// ProductUpdatedEvent is emitted when product details are updated
type ProductUpdatedEvent struct {
	BaseEvent
}

func NewProductUpdatedEvent(aggregateID string) ProductUpdatedEvent {
	return ProductUpdatedEvent{
		BaseEvent: NewBaseEvent(aggregateID, "product.updated"),
	}
}

// ProductActivatedEvent is emitted when a product is activated
type ProductActivatedEvent struct {
	BaseEvent
}

func NewProductActivatedEvent(aggregateID string) ProductActivatedEvent {
	return ProductActivatedEvent{
		BaseEvent: NewBaseEvent(aggregateID, "product.activated"),
	}
}

// ProductDeactivatedEvent is emitted when a product is deactivated
type ProductDeactivatedEvent struct {
	BaseEvent
}

func NewProductDeactivatedEvent(aggregateID string) ProductDeactivatedEvent {
	return ProductDeactivatedEvent{
		BaseEvent: NewBaseEvent(aggregateID, "product.deactivated"),
	}
}

// ProductArchivedEvent is emitted when a product is archived
type ProductArchivedEvent struct {
	BaseEvent
}

func NewProductArchivedEvent(aggregateID string) ProductArchivedEvent {
	return ProductArchivedEvent{
		BaseEvent: NewBaseEvent(aggregateID, "product.archived"),
	}
}

// DiscountAppliedEvent is emitted when a discount is applied to a product
type DiscountAppliedEvent struct {
	BaseEvent
	DiscountPercent int64
	StartDate       int64
	EndDate         int64
}

func NewDiscountAppliedEvent(aggregateID string, discountPercent int64, startDate, endDate int64) DiscountAppliedEvent {
	return DiscountAppliedEvent{
		BaseEvent:       NewBaseEvent(aggregateID, "discount.applied"),
		DiscountPercent: discountPercent,
		StartDate:       startDate,
		EndDate:         endDate,
	}
}

// DiscountRemovedEvent is emitted when a discount is removed from a product
type DiscountRemovedEvent struct {
	BaseEvent
}

func NewDiscountRemovedEvent(aggregateID string) DiscountRemovedEvent {
	return DiscountRemovedEvent{
		BaseEvent: NewBaseEvent(aggregateID, "discount.removed"),
	}
}
