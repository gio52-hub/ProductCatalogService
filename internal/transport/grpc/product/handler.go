package product

import (
	"context"

	"github.com/product-catalog-service/internal/app/product/queries"
	"github.com/product-catalog-service/internal/app/product/usecases"
	pb "github.com/product-catalog-service/proto/product/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler implements the ProductServiceServer interface.
type Handler struct {
	pb.UnimplementedProductServiceServer
	useCases *usecases.ProductUseCases
	queries  *queries.ProductQueries
}

// NewHandler creates a new ProductService gRPC handler.
func NewHandler(useCases *usecases.ProductUseCases, queries *queries.ProductQueries) *Handler {
	return &Handler{
		useCases: useCases,
		queries:  queries,
	}
}

// CreateProduct creates a new product.
func (h *Handler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductReply, error) {
	if err := validateCreateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	appReq := usecases.CreateProductRequest{
		Name:                 req.GetName(),
		Description:          req.GetDescription(),
		Category:             req.GetCategory(),
		BasePriceNumerator:   req.GetBasePrice().GetNumerator(),
		BasePriceDenominator: req.GetBasePrice().GetDenominator(),
	}

	resp, err := h.useCases.CreateProduct(ctx, appReq)
	if err != nil {
		return nil, MapDomainErrorToGRPC(err)
	}

	return &pb.CreateProductReply{
		ProductId: resp.ProductID,
	}, nil
}

// UpdateProduct updates an existing product.
func (h *Handler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.UpdateProductReply, error) {
	if err := validateUpdateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	appReq := usecases.UpdateProductRequest{
		ProductID:   req.GetProductId(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Category:    req.GetCategory(),
	}

	if err := h.useCases.UpdateProduct(ctx, appReq); err != nil {
		return nil, MapDomainErrorToGRPC(err)
	}

	return &pb.UpdateProductReply{}, nil
}

// ActivateProduct activates a product.
func (h *Handler) ActivateProduct(ctx context.Context, req *pb.ActivateProductRequest) (*pb.ActivateProductReply, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	appReq := usecases.ActivateProductRequest{
		ProductID: req.GetProductId(),
	}

	if err := h.useCases.ActivateProduct(ctx, appReq); err != nil {
		return nil, MapDomainErrorToGRPC(err)
	}

	return &pb.ActivateProductReply{}, nil
}

// DeactivateProduct deactivates a product.
func (h *Handler) DeactivateProduct(ctx context.Context, req *pb.DeactivateProductRequest) (*pb.DeactivateProductReply, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	appReq := usecases.DeactivateProductRequest{
		ProductID: req.GetProductId(),
	}

	if err := h.useCases.DeactivateProduct(ctx, appReq); err != nil {
		return nil, MapDomainErrorToGRPC(err)
	}

	return &pb.DeactivateProductReply{}, nil
}

// ArchiveProduct archives a product (soft delete).
func (h *Handler) ArchiveProduct(ctx context.Context, req *pb.ArchiveProductRequest) (*pb.ArchiveProductReply, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	appReq := usecases.ArchiveProductRequest{
		ProductID: req.GetProductId(),
	}

	if err := h.useCases.ArchiveProduct(ctx, appReq); err != nil {
		return nil, MapDomainErrorToGRPC(err)
	}

	return &pb.ArchiveProductReply{}, nil
}

// ApplyDiscount applies a discount to a product.
func (h *Handler) ApplyDiscount(ctx context.Context, req *pb.ApplyDiscountRequest) (*pb.ApplyDiscountReply, error) {
	if err := validateApplyDiscountRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	appReq := usecases.ApplyDiscountRequest{
		ProductID:          req.GetProductId(),
		DiscountPercentage: req.GetDiscountPercentage(),
		StartDate:          req.GetStartDate().AsTime(),
		EndDate:            req.GetEndDate().AsTime(),
	}

	if err := h.useCases.ApplyDiscount(ctx, appReq); err != nil {
		return nil, MapDomainErrorToGRPC(err)
	}

	return &pb.ApplyDiscountReply{}, nil
}

// RemoveDiscount removes a discount from a product.
func (h *Handler) RemoveDiscount(ctx context.Context, req *pb.RemoveDiscountRequest) (*pb.RemoveDiscountReply, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	appReq := usecases.RemoveDiscountRequest{
		ProductID: req.GetProductId(),
	}

	if err := h.useCases.RemoveDiscount(ctx, appReq); err != nil {
		return nil, MapDomainErrorToGRPC(err)
	}

	return &pb.RemoveDiscountReply{}, nil
}

// GetProduct retrieves a product by ID.
func (h *Handler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductReply, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	appReq := queries.GetProductRequest{
		ProductID: req.GetProductId(),
	}

	resp, err := h.queries.GetProduct(ctx, appReq)
	if err != nil {
		return nil, MapDomainErrorToGRPC(err)
	}

	return &pb.GetProductReply{
		Product: MapProductResponseToProto(resp),
	}, nil
}

// ListProducts lists products with optional filters and pagination.
func (h *Handler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsReply, error) {
	appReq := queries.ListProductsRequest{
		Category:   req.GetCategory(),
		Status:     req.GetStatus(),
		ActiveOnly: req.GetActiveOnly(),
		PageSize:   req.GetPageSize(),
		PageToken:  req.GetPageToken(),
	}

	resp, err := h.queries.ListProducts(ctx, appReq)
	if err != nil {
		return nil, MapDomainErrorToGRPC(err)
	}

	return MapListProductsResponseToProto(resp), nil
}
