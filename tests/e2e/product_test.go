// Package e2e contains end-to-end tests
// To run these tests, you need:
// 1. Spanner emulator running on localhost:9010
// 2. Database created and migrations applied
// 3. Run: go test -v ./tests/e2e/
package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"product-catalog-service/internal/app/product/activate_product"
	"product-catalog-service/internal/app/product/apply_discount"
	"product-catalog-service/internal/app/product/create_product"
	"product-catalog-service/internal/app/product/deactivate_product"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/get_product"
	"product-catalog-service/internal/app/product/list_products"
	"product-catalog-service/internal/app/product/repo"
	"product-catalog-service/internal/pkg/clock"
	"product-catalog-service/internal/pkg/committer"
)

const (
	testDatabase = "projects/test-project/instances/test-instance/databases/test-e2e"
)

func setupTest(t *testing.T) (context.Context, *spanner.Client, func()) {
	ctx := context.Background()

	// Create Spanner client for emulator
	client, err := spanner.NewClient(ctx, testDatabase)
	require.NoError(t, err, "Failed to create Spanner client")

	cleanup := func() {
		client.Close()
	}

	return ctx, client, cleanup
}

func TestProductCreationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, client, cleanup := setupTest(t)
	defer cleanup()

	// Setup components
	clk := clock.NewMockClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	committer := committer.NewCommitter(client)
	productRepo := repo.NewProductRepo(client)
	outboxRepo := repo.NewOutboxRepo(client)

	// Create usecase
	createProduct := create_product.NewInteractor(productRepo, outboxRepo, committer, clk)

	// Test: Create product
	req := create_product.Request{
		Name:                 "Test Product",
		Description:          "A test product",
		Category:             "electronics",
		BasePriceNumerator:   1999,
		BasePriceDenominator: 100,
	}

	resp, err := createProduct.Execute(ctx, req)
	require.NoError(t, err, "CreateProduct should succeed")
	assert.NotEmpty(t, resp.ProductID, "Product ID should be returned")

	// Verify: Query returns correct data
	readModel := repo.NewProductReadModel(client)
	getProduct := get_product.NewQuery(readModel)

	getResp, err := getProduct.Execute(ctx, get_product.Request{ProductID: resp.ProductID})
	require.NoError(t, err, "GetProduct should succeed")
	assert.Equal(t, "Test Product", getResp.Product.Name)
	assert.Equal(t, "electronics", getResp.Product.Category)
	assert.Equal(t, int64(1999), getResp.Product.BasePriceNumerator)
	assert.Equal(t, int64(100), getResp.Product.BasePriceDenominator)
	assert.Equal(t, "active", getResp.Product.Status)

	t.Logf("✓ Product created and retrieved successfully")
}

func TestDiscountApplicationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, client, cleanup := setupTest(t)
	defer cleanup()

	// Setup
	fixedTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	clk := clock.NewMockClock(fixedTime)
	committer := committer.NewCommitter(client)
	productRepo := repo.NewProductRepo(client)
	outboxRepo := repo.NewOutboxRepo(client)
	enricher := &testEventEnricher{}

	// Create product first
	createProduct := create_product.NewInteractor(productRepo, outboxRepo, committer, clk)
	createReq := create_product.Request{
		Name:                 "Premium Product",
		Description:          "High quality product",
		Category:             "electronics",
		BasePriceNumerator:   10000,
		BasePriceDenominator: 100,
	}
	createResp, err := createProduct.Execute(ctx, createReq)
	require.NoError(t, err)

	// Apply discount
	applyDiscount := apply_discount.NewInteractor(productRepo, productRepo, outboxRepo, committer, clk, enricher)

	startTime := fixedTime.Add(-time.Hour).Unix()
	endTime := fixedTime.Add(24 * time.Hour).Unix()

	applyReq := apply_discount.Request{
		ProductID:       createResp.ProductID,
		DiscountPercent: 20,
		DiscountStartSec: startTime,
		DiscountEndSec:   endTime,
	}

	_, err = applyDiscount.Execute(ctx, applyReq)
	require.NoError(t, err, "ApplyDiscount should succeed")

	// Verify: Effective price is calculated correctly
	readModel := repo.NewProductReadModel(client)
	getProduct := get_product.NewQuery(readModel)

	getResp, err := getProduct.Execute(ctx, get_product.Request{ProductID: createResp.ProductID})
	require.NoError(t, err)

	// Price: 100.00 - 20% = 80.00
	assert.True(t, getResp.Product.HasDiscount)
	assert.Equal(t, int64(20), *getResp.Product.DiscountPercent)
	assert.Equal(t, int64(10000*80), getResp.Product.EffectivePriceNumerator) // 10000 * 0.80 = 8000
	assert.Equal(t, int64(100*100), getResp.Product.EffectivePriceDenominator)

	t.Logf("✓ Discount applied and effective price calculated correctly")
}

func TestBusinessRuleValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, client, cleanup := setupTest(t)
	defer cleanup()

	// Setup
	fixedTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	clk := clock.NewMockClock(fixedTime)
	committer := committer.NewCommitter(client)
	productRepo := repo.NewProductRepo(client)
	outboxRepo := repo.NewOutboxRepo(client)
	enricher := &testEventEnricher{}

	// Create product
	createProduct := create_product.NewInteractor(productRepo, outboxRepo, committer, clk)
	createReq := create_product.Request{
		Name:                 "Test Product",
		Description:          "Description",
		Category:             "category",
		BasePriceNumerator:   100,
		BasePriceDenominator: 1,
	}
	createResp, err := createProduct.Execute(ctx, createReq)
	require.NoError(t, err)

	// Test: Cannot apply discount to inactive product
	deactivate := deactivate_product.NewInteractor(productRepo, productRepo, outboxRepo, committer, clk, enricher)
	_, err = deactivate.Execute(ctx, deactivate_product.Request{ProductID: createResp.ProductID})
	require.NoError(t, err)

	applyDiscount := apply_discount.NewInteractor(productRepo, productRepo, outboxRepo, committer, clk, enricher)

	applyReq := apply_discount.Request{
		ProductID:       createResp.ProductID,
		DiscountPercent: 10,
		DiscountStartSec: fixedTime.Add(-time.Hour).Unix(),
		DiscountEndSec:   fixedTime.Add(24 * time.Hour).Unix(),
	}

	_, err = applyDiscount.Execute(ctx, applyReq)
	assert.ErrorIs(t, err, domain.ErrProductNotActive, "Should not allow discount on inactive product")

	t.Logf("✓ Business rule validated correctly")
}

func TestProductActivationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, client, cleanup := setupTest(t)
	defer cleanup()

	// Setup
	fixedTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	clk := clock.NewMockClock(fixedTime)
	committer := committer.NewCommitter(client)
	productRepo := repo.NewProductRepo(client)
	outboxRepo := repo.NewOutboxRepo(client)
	enricher := &testEventEnricher{}

	// Create inactive product
	createProduct := create_product.NewInteractor(productRepo, outboxRepo, committer, clk)
	createReq := create_product.Request{
		Name:                 "New Product",
		Description:          "Description",
		Category:             "category",
		BasePriceNumerator:   100,
		BasePriceDenominator: 1,
	}
	createResp, err := createProduct.Execute(ctx, createReq)
	require.NoError(t, err)

	// Deactivate
	deactivate := deactivate_product.NewInteractor(productRepo, productRepo, outboxRepo, committer, clk, enricher)
	_, err = deactivate.Execute(ctx, deactivate_product.Request{ProductID: createResp.ProductID})
	require.NoError(t, err)

	// Activate
	activate := activate_product.NewInteractor(productRepo, productRepo, outboxRepo, committer, clk, enricher)
	_, err = activate.Execute(ctx, activate_product.Request{ProductID: createResp.ProductID})
	require.NoError(t, err)

	// Verify status
	readModel := repo.NewProductReadModel(client)
	getProduct := get_product.NewQuery(readModel)

	getResp, err := getProduct.Execute(ctx, get_product.Request{ProductID: createResp.ProductID})
	require.NoError(t, err)
	assert.Equal(t, "active", getResp.Product.Status)

	t.Logf("✓ Product activation flow working correctly")
}

func TestListProductsWithPagination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, client, cleanup := setupTest(t)
	defer cleanup()

	// Setup
	clk := clock.NewMockClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	committer := committer.NewCommitter(client)
	productRepo := repo.NewProductRepo(client)
	outboxRepo := repo.NewOutboxRepo(client)

	// Create multiple products
	createProduct := create_product.NewInteractor(productRepo, outboxRepo, committer, clk)

	for i := 0; i < 5; i++ {
		req := create_product.Request{
			Name:                 "Product " + string(rune('A'+i)),
			Description:          "Test product",
			Category:             "test",
			BasePriceNumerator:   100,
			BasePriceDenominator: 1,
		}
		_, err := createProduct.Execute(ctx, req)
		require.NoError(t, err)
	}

	// List products with pagination
	readModel := repo.NewProductReadModel(client)
	listProducts := list_products.NewQuery(readModel)

	listResp, err := listProducts.Execute(ctx, list_products.Request{
		PageSize: 3,
	})
	require.NoError(t, err)
	assert.Len(t, listResp.Products, 3, "Should return page size of 3")
	assert.NotEmpty(t, listResp.NextPageToken, "Should have next page token")

	// Get next page
	page2Resp, err := listProducts.Execute(ctx, list_products.Request{
		PageSize:  3,
		PageToken: listResp.NextPageToken,
	})
	require.NoError(t, err)
	assert.Len(t, page2Resp.Products, 2, "Should return remaining 2 products")
	assert.Empty(t, page2Resp.NextPageToken, "Should have no more pages")

	t.Logf("✓ Pagination working correctly")
}

// Test event enricher for usecases
type testEventEnricher struct{}

func (e *testEventEnricher) EnrichEvent(event interface{}) create_product.OutboxEvent {
	// Simple enricher - just return a basic outbox event
	return create_product.OutboxEvent{
		EventID:     "test-event-id",
		EventType:   "test.event",
		AggregateID: "test-aggregate",
		Payload:     make(map[string]interface{}),
	}
}

// TestMain is the entry point for running tests directly
func TestMain(m *testing.M) {
	// Check if we should skip E2E tests
	if os.Getenv("SPANNER_EMULATOR_HOST") == "" {
		m.Log("E2E tests skipped - set SPANNER_EMULATOR_HOST to run")
		return
	}

	os.Exit(m.Run())
}
