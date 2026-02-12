package repo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc/codes"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/models/m_product"
)

// ProductReadModel implements ProductReadModel for Spanner
type ProductReadModel struct {
	client *spanner.Client
}

// NewProductReadModel creates a new Spanner product read model
func NewProductReadModel(client *spanner.Client) *ProductReadModel {
	return &ProductReadModel{
		client: client,
	}
}

// GetProduct retrieves a product by ID with effective price calculated
func (r *ProductReadModel) GetProduct(ctx context.Context, productID string) (*contracts.ProductDTO, error) {
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
		},
	)

	if err != nil {
		// Check for not found error
		if spanner.ErrCode(err) == codes.NotFound {
			return nil, fmt.Errorf("product not found: %s", productID)
		}
		return nil, fmt.Errorf("failed to read product: %w", err)
	}

	var (
		productIDVal            string
		name                    string
		description             string
		category                string
		basePriceNum            int64
		basePriceDenom           int64
		discountPercent         *int64
		discountStart           *time.Time
		discountEnd             *time.Time
		status                  string
		createdAt               time.Time
		updatedAt               time.Time
	)

	if err := row.Columns(
		&productIDVal,
		&name,
		&description,
		&category,
		&basePriceNum,
		&basePriceDenom,
		&discountPercent,
		&discountStart,
		&discountEnd,
		&status,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, fmt.Errorf("failed to parse product row: %w", err)
	}

	dto := &contracts.ProductDTO{
		ProductID:               productIDVal,
		Name:                    name,
		Description:             description,
		Category:                category,
		BasePriceNumerator:      basePriceNum,
		BasePriceDenominator:    basePriceDenom,
		Status:                  status,
		CreatedAtSec:            createdAt.Unix(),
		UpdatedAtSec:            updatedAt.Unix(),
		EffectivePriceNumerator: basePriceNum,
		EffectivePriceDenominator: basePriceDenom,
	}

	// Calculate effective price if discount is active
	now := time.Now()
	if discountPercent != nil && discountStart != nil && discountEnd != nil {
		if (now.Equal(*discountStart) || now.After(*discountStart)) &&
		   (now.Before(*discountEnd) || now.Equal(*discountEnd)) {
			dto.HasDiscount = true
			dto.DiscountPercent = discountPercent
			dto.DiscountStartDate = &[]int64{discountStart.Unix()}[0]
			dto.DiscountEndDate = &[]int64{discountEnd.Unix()}[0]

			// Calculate effective price: base * (1 - discount/100)
			// effectiveNum = baseNum * (100 - discountPercent)
			// effectiveDenom = baseDenom * 100
			discountFactor := 100 - *discountPercent
			dto.EffectivePriceNumerator = basePriceNum * discountFactor
			dto.EffectivePriceDenominator = basePriceDenom * 100
		}
	}

	return dto, nil
}

// ListProducts retrieves a paginated list of products
func (r *ProductReadModel) ListProducts(ctx context.Context, filter contracts.ListProductsFilter) (*contracts.PaginatedProductsDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Set default page size
	pageSize := filter.PageSize
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 50
	}

	// Build query
	stmt := spanner.NewStatement(`
		SELECT
			product_id, name, description, category,
			base_price_numerator, base_price_denominator,
			discount_percent, discount_start_date, discount_end_date,
			status, created_at, updated_at
		FROM products
		WHERE @status IS NULL OR status = @status
			AND (@category IS NULL OR category = @category)
		ORDER BY product_id
		LIMIT @limit
	`)

	params := map[string]interface{}{
		"status":   filter.Status,
		"category": filter.Category,
		"limit":    pageSize + 1, // Fetch one extra to determine if there's a next page
	}

	stmt.Params = params

	txn := r.client.ReadOnlyTransaction()
	defer txn.Close()

	iter := txn.Query(ctx, stmt)
	defer iter.Stop()

	var products []*contracts.ProductDTO

	for {
		row, err := iter.Next()
		if err != nil {
			if spanner.ErrCode(err) == codes.NotFound {
				break
			}
			return nil, fmt.Errorf("failed to iterate products: %w", err)
		}

		var (
			productIDVal            string
			name                    string
			description             string
			category                string
			basePriceNum            int64
			basePriceDenom           int64
			discountPercent         *int64
			discountStart           *time.Time
			discountEnd             *time.Time
			status                  string
			createdAt               time.Time
			updatedAt               time.Time
		)

		if err := row.Columns(
			&productIDVal,
			&name,
			&description,
			&category,
			&basePriceNum,
			&basePriceDenom,
			&discountPercent,
			&discountStart,
			&discountEnd,
			&status,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to parse product row: %w", err)
		}

		dto := &contracts.ProductDTO{
			ProductID:               productIDVal,
			Name:                    name,
			Description:             description,
			Category:                category,
			BasePriceNumerator:      basePriceNum,
			BasePriceDenominator:    basePriceDenom,
			Status:                  status,
			CreatedAtSec:            createdAt.Unix(),
			UpdatedAtSec:            updatedAt.Unix(),
			EffectivePriceNumerator: basePriceNum,
			EffectivePriceDenominator: basePriceDenom,
		}

		// Calculate effective price if discount is active
		now := time.Now()
		if discountPercent != nil && discountStart != nil && discountEnd != nil {
			if (now.Equal(*discountStart) || now.After(*discountStart)) &&
			   (now.Before(*discountEnd) || now.Equal(*discountEnd)) {
				dto.HasDiscount = true
				dto.DiscountPercent = discountPercent
				dto.DiscountStartDate = &[]int64{discountStart.Unix()}[0]
				dto.DiscountEndDate = &[]int64{discountEnd.Unix()}[0]

				// Calculate effective price
				discountFactor := 100 - *discountPercent
				dto.EffectivePriceNumerator = basePriceNum * discountFactor
				dto.EffectivePriceDenominator = basePriceDenom * 100
			}
		}

		products = append(products, dto)
	}

	// Check if there's a next page
	var nextPageToken string
	if len(products) > pageSize {
		products = products[:pageSize]
		lastProduct := products[len(products)-1]
		nextPageToken = encodePageToken(lastProduct.ProductID)
	}

	return &contracts.PaginatedProductsDTO{
		Products:      products,
		NextPageToken: nextPageToken,
	}, nil
}

const iteratorDone = "spanner: iterator done"

func encodePageToken(productID string) string {
	return base64.StdEncoding.EncodeToString([]byte(productID))
}

func decodePageToken(token string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Helper for pagination - you can extend this to decode tokens
func init() {
	// Register custom JSON encoders if needed
	json.Marshal(struct{}{})
}
