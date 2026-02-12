package domain

import (
	"time"
)

// ProductStatus represents the status of a product
type ProductStatus string

const (
	ProductStatusActive    ProductStatus = "active"
	ProductStatusInactive  ProductStatus = "inactive"
	ProductStatusArchived  ProductStatus = "archived"
)

// Product is the aggregate root for products
type Product struct {
	id          string
	name        string
	description string
	category    string
	basePrice   *Money
	discount    *Discount
	status      ProductStatus
	createdAt   time.Time
	updatedAt   time.Time
	archivedAt  *time.Time
	changes     *ChangeTracker
	events      []DomainEvent
	version     int
}

// NewProduct creates a new product
func NewProduct(id, name, description, category string, basePrice *Money, now time.Time) (*Product, error) {
	if id == "" {
		return nil, ErrInvalidName
	}
	if name == "" {
		return nil, ErrInvalidName
	}
	if category == "" {
		return nil, ErrInvalidCategory
	}
	if basePrice == nil {
		return nil, ErrInvalidPrice
	}

	p := &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		status:      ProductStatusActive,
		createdAt:   now,
		updatedAt:   now,
		changes:     NewChangeTracker(),
		events:      make([]DomainEvent, 0),
		version:     0,
	}

	// Mark all fields as dirty for new product
	p.changes.MarkDirty(FieldID)
	p.changes.MarkDirty(FieldName)
	p.changes.MarkDirty(FieldDescription)
	p.changes.MarkDirty(FieldCategory)
	p.changes.MarkDirty(FieldBasePrice)
	p.changes.MarkDirty(FieldStatus)

	// Record creation event
	p.recordEvent(NewProductCreatedEvent(
		id,
		name,
		category,
		basePrice.Numerator(),
		basePrice.Denominator(),
	))

	return p, nil
}

// ReconstructProduct reconstructs a product from persistence
func ReconstructProduct(
	id, name, description, category string,
	basePriceNum, basePriceDenom int64,
	discountPercent int64,
	discountStart, discountEnd time.Time,
	status string,
	createdAt, updatedAt time.Time,
	archivedAt *time.Time,
	version int,
) (*Product, error) {
	basePrice, err := NewMoney(basePriceNum, basePriceDenom)
	if err != nil {
		return nil, err
	}

	var discount *Discount
	if !discountStart.IsZero() && !discountEnd.IsZero() {
		discount, err = NewDiscount(discountPercent, discountStart, discountEnd)
		if err != nil {
			return nil, err
		}
	}

	return &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		discount:    discount,
		status:      ProductStatus(status),
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		archivedAt:  archivedAt,
		changes:     NewChangeTracker(),
		events:      make([]DomainEvent, 0),
		version:     version,
	}, nil
}

// Accessor methods

func (p *Product) ID() string                 { return p.id }
func (p *Product) Name() string               { return p.name }
func (p *Product) Description() string       { return p.description }
func (p *Product) Category() string           { return p.category }
func (p *Product) BasePrice() *Money          { return p.basePrice }
func (p *Product) Discount() *Discount        { return p.discount }
func (p *Product) Status() ProductStatus      { return p.status }
func (p *Product) CreatedAt() time.Time       { return p.createdAt }
func (p *Product) UpdatedAt() time.Time       { return p.updatedAt }
func (p *Product) ArchivedAt() *time.Time      { return p.archivedAt }
func (p *Product) Changes() *ChangeTracker    { return p.changes }
func (p *Product) Version() int               { return p.version }

// DomainEvents returns all recorded events
func (p *Product) DomainEvents() []DomainEvent {
	return p.events
}

// ClearEvents clears all recorded events
func (p *Product) ClearEvents() {
	p.events = make([]DomainEvent, 0)
}

// Business methods

// UpdateDetails updates the product's name, description, and category
func (p *Product) UpdateDetails(name, description, category string, now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductIsArchived
	}

	if name == "" {
		return ErrInvalidName
	}
	if category == "" {
		return ErrInvalidCategory
	}

	updated := false
	if p.name != name {
		p.name = name
		p.changes.MarkDirty(FieldName)
		updated = true
	}

	if p.description != description {
		p.description = description
		p.changes.MarkDirty(FieldDescription)
		updated = true
	}

	if p.category != category {
		p.category = category
		p.changes.MarkDirty(FieldCategory)
		updated = true
	}

	if updated {
		p.updatedAt = now
		p.changes.MarkDirty(FieldStatus) // Status field includes updated_at
		p.recordEvent(NewProductUpdatedEvent(p.id))
	}

	return nil
}

// Activate activates the product
func (p *Product) Activate(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductIsArchived
	}

	if p.status == ProductStatusActive {
		return ErrProductAlreadyActive
	}

	p.status = ProductStatusActive
	p.updatedAt = now
	p.changes.MarkDirty(FieldStatus)
	p.recordEvent(NewProductActivatedEvent(p.id))

	return nil
}

// Deactivate deactivates the product
func (p *Product) Deactivate(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductIsArchived
	}

	if p.status == ProductStatusInactive {
		return nil // Already inactive
	}

	p.status = ProductStatusInactive
	p.updatedAt = now
	p.changes.MarkDirty(FieldStatus)
	p.recordEvent(NewProductDeactivatedEvent(p.id))

	return nil
}

// ApplyDiscount applies a discount to the product
func (p *Product) ApplyDiscount(discount *Discount, now time.Time) error {
	if p.status != ProductStatusActive {
		return ErrProductNotActive
	}

	if !discount.IsValidAt(now) {
		return ErrInvalidDiscountPeriod
	}

	p.discount = discount
	p.updatedAt = now
	p.changes.MarkDirty(FieldDiscount)
	p.changes.MarkDirty(FieldStatus) // Status field includes updated_at

	p.recordEvent(NewDiscountAppliedEvent(
		p.id,
		discount.Percentage(),
		discount.StartDate().Unix(),
		discount.EndDate().Unix(),
	))

	return nil
}

// RemoveDiscount removes the active discount
func (p *Product) RemoveDiscount(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductIsArchived
	}

	if p.discount == nil {
		return ErrNoActiveDiscount
	}

	p.discount = nil
	p.updatedAt = now
	p.changes.MarkDirty(FieldDiscount)
	p.changes.MarkDirty(FieldStatus)

	p.recordEvent(NewDiscountRemovedEvent(p.id))

	return nil
}

// Archive archives the product (soft delete)
func (p *Product) Archive(now time.Time) error {
	if p.status == ProductStatusArchived {
		return nil // Already archived
	}

	p.status = ProductStatusArchived
	p.updatedAt = now
	p.archivedAt = &now
	p.changes.MarkDirty(FieldStatus)
	p.changes.MarkDirty(FieldArchivedAt)

	p.recordEvent(NewProductArchivedEvent(p.id))

	return nil
}

// EffectivePrice calculates the price after applying any active discount
func (p *Product) EffectivePrice(now time.Time) (*Money, error) {
	if p.discount == nil || !p.discount.IsActiveAt(now) {
		return p.basePrice, nil
	}

	return p.basePrice.ApplyPercentage(p.discount.Percentage())
}

// IncrementVersion increments the version for optimistic locking
func (p *Product) IncrementVersion() {
	p.version++
}

// recordEvent adds a domain event
func (p *Product) recordEvent(event DomainEvent) {
	p.events = append(p.events, event)
}
