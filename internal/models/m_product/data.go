package m_product

import (
	"time"
)

// Product represents a database row in the products table
type Product struct {
	ProductID            string
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
	DiscountPercent      *int64
	DiscountStartDate    *time.Time
	DiscountEndDate      *time.Time
	Status               string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ArchivedAt           *time.Time
}

// ToMap converts the product to a map for Spanner mutation
func (p *Product) ToMap() map[string]interface{} {
	return map[string]interface{}{
		ProductID:            p.ProductID,
		Name:                 p.Name,
		Description:          p.Description,
		Category:             p.Category,
		BasePriceNumerator:   p.BasePriceNumerator,
		BasePriceDenominator: p.BasePriceDenominator,
		DiscountPercent:      p.DiscountPercent,
		DiscountStartDate:    p.DiscountStartDate,
		DiscountEndDate:      p.DiscountEndDate,
		Status:               p.Status,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
		ArchivedAt:           p.ArchivedAt,
	}
}
