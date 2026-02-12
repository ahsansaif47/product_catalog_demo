package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc/codes"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/models/m_outbox"
	"product-catalog-service/internal/models/m_product"
)

type spannerContext = context.Context

// ProductRepo implements ProductRepository for Spanner
type ProductRepo struct {
	client *spanner.Client
}

// NewProductRepo creates a new Spanner product repository
func NewProductRepo(client *spanner.Client) *ProductRepo {
	return &ProductRepo{
		client: client,
	}
}

// InsertMut returns a mutation to insert a product
func (r *ProductRepo) InsertMut(product *domain.Product) *spanner.Mutation {
	p := r.domainToModel(product)

	mutation := spanner.InsertOrUpdateMap(m_product.Table, p.ToMap())
	return mutation
}

// UpdateMut returns a mutation to update a product (targeted by change tracker)
func (r *ProductRepo) UpdateMut(product *domain.Product) *spanner.Mutation {
	updates := make(map[string]interface{})

	// Only update fields that have changed
	if product.Changes().Dirty(domain.FieldName) {
		updates[m_product.Name] = product.Name()
	}

	if product.Changes().Dirty(domain.FieldDescription) {
		updates[m_product.Description] = product.Description()
	}

	if product.Changes().Dirty(domain.FieldCategory) {
		updates[m_product.Category] = product.Category()
	}

	if product.Changes().Dirty(domain.FieldBasePrice) {
		updates[m_product.BasePriceNumerator] = product.BasePrice().Numerator()
		updates[m_product.BasePriceDenominator] = product.BasePrice().Denominator()
	}

	if product.Changes().Dirty(domain.FieldDiscount) {
		if d := product.Discount(); d != nil {
			updates[m_product.DiscountPercent] = d.Percentage()
			updates[m_product.DiscountStartDate] = d.StartDate()
			updates[m_product.DiscountEndDate] = d.EndDate()
		} else {
			updates[m_product.DiscountPercent] = nil
			updates[m_product.DiscountStartDate] = nil
			updates[m_product.DiscountEndDate] = nil
		}
	}

	if product.Changes().Dirty(domain.FieldStatus) || product.Changes().HasChanges() {
		updates[m_product.Status] = string(product.Status())
		updates[m_product.UpdatedAt] = time.Now()
	}

	if product.Changes().Dirty(domain.FieldArchivedAt) {
		updates[m_product.ArchivedAt] = product.ArchivedAt()
	}

	if len(updates) == 0 {
		return nil // No changes to apply
	}

	mutation := spanner.UpdateMap(m_product.Table, updates)
	return mutation
}

// FindByID retrieves a product by ID
func (r *ProductRepo) FindByID(ctx spannerContext, productID string) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	txn := r.client.ReadOnlyTransaction()
	defer txn.Close()

	row, err := txn.ReadRow(ctx, m_product.Table, spanner.Key{productID},
		[]string{
			m_product.ProductID,
			m_product.Name,
			m_product.Description,
			m_product.Category,
			m_product.BasePriceNumerator,
			m_product.BasePriceDenominator,
			m_product.DiscountPercent,
			m_product.DiscountStartDate,
			m_product.DiscountEndDate,
			m_product.Status,
			m_product.CreatedAt,
			m_product.UpdatedAt,
			m_product.ArchivedAt,
		},
	)

	if err != nil {
		// Check for not found error
		if spanner.ErrCode(err) == codes.NotFound {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to read product: %w", err)
	}

	var p m_product.Product
	var discountPercent *int64
	var discountStart, discountEnd, archivedAt *time.Time

	if err := row.Columns(
		&p.ProductID,
		&p.Name,
		&p.Description,
		&p.Category,
		&p.BasePriceNumerator,
		&p.BasePriceDenominator,
		&discountPercent,
		&discountStart,
		&discountEnd,
		&p.Status,
		&p.CreatedAt,
		&p.UpdatedAt,
		&archivedAt,
	); err != nil {
		return nil, fmt.Errorf("failed to parse product row: %w", err)
	}

	p.DiscountPercent = discountPercent
	p.DiscountStartDate = discountStart
	p.DiscountEndDate = discountEnd
	p.ArchivedAt = archivedAt

	return r.modelToDomain(&p)
}

// Exists checks if a product exists
func (r *ProductRepo) Exists(ctx spannerContext, productID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	txn := r.client.ReadOnlyTransaction()
	defer txn.Close()

	_, err := txn.ReadRow(ctx, m_product.Table, spanner.Key{productID}, []string{m_product.ProductID})
	if err != nil {
		// Check for not found error
		if spanner.ErrCode(err) == codes.NotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check product existence: %w", err)
	}

	return true, nil
}

// OutboxRepo implements OutboxRepository for Spanner
type OutboxRepo struct {
	client *spanner.Client
}

// NewOutboxRepo creates a new Spanner outbox repository
func NewOutboxRepo(client *spanner.Client) *OutboxRepo {
	return &OutboxRepo{
		client: client,
	}
}

// InsertMut returns a mutation to insert an outbox event
func (r *OutboxRepo) InsertMut(event contracts.OutboxEvent) *spanner.Mutation {
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return nil
	}

	outboxEvent := &m_outbox.OutboxEvent{
		EventID:     event.EventID,
		EventType:   event.EventType,
		AggregateID: event.AggregateID,
		Payload:     string(payloadJSON),
		Status:      m_outbox.StatusPending,
		CreatedAt:   time.Now(),
		ProcessedAt: nil,
	}

	mutation := spanner.InsertOrUpdateMap(m_outbox.Table, outboxEvent.ToMap())
	return mutation
}

// Helper methods

func (r *ProductRepo) domainToModel(product *domain.Product) *m_product.Product {
	p := &m_product.Product{
		ProductID:            product.ID(),
		Name:                 product.Name(),
		Description:          product.Description(),
		Category:             product.Category(),
		BasePriceNumerator:   product.BasePrice().Numerator(),
		BasePriceDenominator: product.BasePrice().Denominator(),
		Status:               string(product.Status()),
		CreatedAt:            product.CreatedAt(),
		UpdatedAt:            product.UpdatedAt(),
		ArchivedAt:           product.ArchivedAt(),
	}

	if d := product.Discount(); d != nil {
		percent := d.Percentage()
		p.DiscountPercent = &percent
		p.DiscountStartDate = &[]time.Time{d.StartDate()}[0]
		p.DiscountEndDate = &[]time.Time{d.EndDate()}[0]
	}

	return p
}

func (r *ProductRepo) modelToDomain(p *m_product.Product) (*domain.Product, error) {
	var discountPercent int64
	var discountStart, discountEnd time.Time

	hasDiscount := p.DiscountPercent != nil
	if hasDiscount {
		discountPercent = *p.DiscountPercent
		if p.DiscountStartDate != nil {
			discountStart = *p.DiscountStartDate
		}
		if p.DiscountEndDate != nil {
			discountEnd = *p.DiscountEndDate
		}
	}

	return domain.ReconstructProduct(
		p.ProductID,
		p.Name,
		p.Description,
		p.Category,
		p.BasePriceNumerator,
		p.BasePriceDenominator,
		discountPercent,
		discountStart,
		discountEnd,
		p.Status,
		p.CreatedAt,
		p.UpdatedAt,
		p.ArchivedAt,
		0, // version
	)
}
