package product

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"product-catalog-service/internal/app/product/queries/get_product"
	"product-catalog-service/internal/app/product/queries/list_products"
	"product-catalog-service/internal/app/product/usecases/activate_product"
	"product-catalog-service/internal/app/product/usecases/apply_discount"
	"product-catalog-service/internal/app/product/usecases/archive_product"
	"product-catalog-service/internal/app/product/usecases/create_product"
	"product-catalog-service/internal/app/product/usecases/deactivate_product"
	"product-catalog-service/internal/app/product/usecases/remove_discount"
	"product-catalog-service/internal/app/product/usecases/update_product"
	productv1 "product-catalog-service/proto/product/v1"
)

// Handlers contains all the product usecase handlers
type Handlers struct {
	createProduct     *create_product.Interactor
	updateProduct     *update_product.Interactor
	activateProduct   *activate_product.Interactor
	deactivateProduct *deactivate_product.Interactor
	applyDiscount     *apply_discount.Interactor
	removeDiscount    *remove_discount.Interactor
	archiveProduct    *archive_product.Interactor
	getProduct        *get_product.Query
	listProducts      *list_products.Query
}

// NewHandlers creates a new product handlers instance
func NewHandlers(
	createProduct *create_product.Interactor,
	updateProduct *update_product.Interactor,
	activateProduct *activate_product.Interactor,
	deactivateProduct *deactivate_product.Interactor,
	applyDiscount *apply_discount.Interactor,
	removeDiscount *remove_discount.Interactor,
	archiveProduct *archive_product.Interactor,
	getProduct *get_product.Query,
	listProducts *list_products.Query,
) *Handlers {
	return &Handlers{
		createProduct:     createProduct,
		updateProduct:     updateProduct,
		activateProduct:   activateProduct,
		deactivateProduct: deactivateProduct,
		applyDiscount:     applyDiscount,
		removeDiscount:    removeDiscount,
		archiveProduct:    archiveProduct,
		getProduct:        getProduct,
		listProducts:      listProducts,
	}
}

// Handler implements the gRPC ProductService server
type Handler struct {
	handlers *Handlers
	productv1.UnimplementedProductServiceServer
}

// NewHandler creates a new gRPC handler
func NewHandler(handlers *Handlers) *Handler {
	return &Handler{
		handlers: handlers,
	}
}

// CreateProduct handles the CreateProduct RPC
func (h *Handler) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductReply, error) {
	// Validate proto request
	if err := validateCreateProductRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Map proto to application request
	appReq := create_product.Request{
		Name:                 req.Name,
		Description:          req.Description,
		Category:             req.Category,
		BasePriceNumerator:   req.BasePriceNumerator,
		BasePriceDenominator: req.BasePriceDenominator,
	}

	// Execute usecase
	resp, err := h.handlers.createProduct.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &productv1.CreateProductReply{
		ProductId: resp.ProductID,
	}, nil
}

// UpdateProduct handles the UpdateProduct RPC
func (h *Handler) UpdateProduct(ctx context.Context, req *productv1.UpdateProductRequest) (*productv1.UpdateProductReply, error) {
	if err := validateUpdateProductRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	appReq := update_product.Request{
		ProductID:   req.ProductId,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
	}

	_, err := h.handlers.updateProduct.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &productv1.UpdateProductReply{}, nil
}

// ActivateProduct handles the ActivateProduct RPC
func (h *Handler) ActivateProduct(ctx context.Context, req *productv1.ActivateProductRequest) (*productv1.ActivateProductReply, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	appReq := activate_product.Request{
		ProductID: req.ProductId,
	}

	_, err := h.handlers.activateProduct.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &productv1.ActivateProductReply{}, nil
}

// DeactivateProduct handles the DeactivateProduct RPC
func (h *Handler) DeactivateProduct(ctx context.Context, req *productv1.DeactivateProductRequest) (*productv1.DeactivateProductReply, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	appReq := deactivate_product.Request{
		ProductID: req.ProductId,
	}

	_, err := h.handlers.deactivateProduct.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &productv1.DeactivateProductReply{}, nil
}

// ApplyDiscount handles the ApplyDiscount RPC
func (h *Handler) ApplyDiscount(ctx context.Context, req *productv1.ApplyDiscountRequest) (*productv1.ApplyDiscountReply, error) {
	if err := validateApplyDiscountRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	appReq := apply_discount.Request{
		ProductID:        req.ProductId,
		DiscountPercent:  req.DiscountPercent,
		DiscountStartSec: req.StartDateSeconds,
		DiscountEndSec:   req.EndDateSeconds,
	}

	_, err := h.handlers.applyDiscount.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &productv1.ApplyDiscountReply{}, nil
}

// RemoveDiscount handles the RemoveDiscount RPC
func (h *Handler) RemoveDiscount(ctx context.Context, req *productv1.RemoveDiscountRequest) (*productv1.RemoveDiscountReply, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	appReq := remove_discount.Request{
		ProductID: req.ProductId,
	}

	_, err := h.handlers.removeDiscount.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &productv1.RemoveDiscountReply{}, nil
}

// ArchiveProduct handles the ArchiveProduct RPC
func (h *Handler) ArchiveProduct(ctx context.Context, req *productv1.ArchiveProductRequest) (*productv1.ArchiveProductReply, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	appReq := archive_product.Request{
		ProductID: req.ProductId,
	}

	_, err := h.handlers.archiveProduct.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &productv1.ArchiveProductReply{}, nil
}

// GetProduct handles the GetProduct RPC
func (h *Handler) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductReply, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	appReq := get_product.Request{
		ProductID: req.ProductId,
	}

	resp, err := h.handlers.getProduct.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &productv1.GetProductReply{
		Product: dtoToProtoProduct(resp.Product),
	}, nil
}

// ListProducts handles the ListProducts RPC
func (h *Handler) ListProducts(ctx context.Context, req *productv1.ListProductsRequest) (*productv1.ListProductsReply, error) {
	appReq := list_products.Request{
		Category:  req.Category,
		PageSize:  int(req.PageSize),
		PageToken: req.PageToken,
		Status:    "", // Default to empty to return all statuses
	}

	resp, err := h.handlers.listProducts.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	products := make([]*productv1.Product, len(resp.Products))
	for i, p := range resp.Products {
		products[i] = dtoToProtoProduct(p)
	}

	return &productv1.ListProductsReply{
		Products:      products,
		NextPageToken: resp.NextPageToken,
	}, nil
}
