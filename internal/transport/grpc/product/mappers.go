package product

import (
	productv1 "product-catalog-service/proto/product/v1"
	"product-catalog-service/internal/app/product/contracts"
)

// dtoToProtoProduct converts a ProductDTO to a proto Product
func dtoToProtoProduct(dto *contracts.ProductDTO) *productv1.Product {
	p := &productv1.Product{
		ProductId: dto.ProductID,
		Name:      dto.Name,
		Description: dto.Description,
		Category:  dto.Category,
		BasePrice: &productv1.Money{
			Numerator:   dto.BasePriceNumerator,
			Denominator: dto.BasePriceDenominator,
		},
		EffectivePrice: &productv1.Money{
			Numerator:   dto.EffectivePriceNumerator,
			Denominator: dto.EffectivePriceDenominator,
		},
		Status:          dto.Status,
		CreatedAtSeconds: dto.CreatedAtSec,
		UpdatedAtSeconds: dto.UpdatedAtSec,
	}

	if dto.HasDiscount {
		p.Discount = &productv1.Discount{
			Percent:         *dto.DiscountPercent,
			StartDateSeconds: *dto.DiscountStartDate,
			EndDateSeconds:   *dto.DiscountEndDate,
		}
	}

	return p
}
