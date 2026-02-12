package services

import (
	"product-catalog-service/internal/app/product/domain"
	"time"
)

// PricingCalculator is a domain service for calculating prices
type PricingCalculator struct{}

// NewPricingCalculator creates a new pricing calculator
func NewPricingCalculator() *PricingCalculator {
	return &PricingCalculator{}
}

// CalculateEffectivePrice calculates the effective price for a product at a given time
func (pc *PricingCalculator) CalculateEffectivePrice(product *domain.Product, now time.Time) (*domain.Money, error) {
	return product.EffectivePrice(now)
}

// CalculateEffectivePrices calculates effective prices for multiple products
func (pc *PricingCalculator) CalculateEffectivePrices(products []*domain.Product, now time.Time) (map[string]*domain.Money, error) {
	result := make(map[string]*domain.Money)

	for _, product := range products {
		price, err := product.EffectivePrice(now)
		if err != nil {
			return nil, err
		}
		result[product.ID()] = price
	}

	return result, nil
}
