package usecase

import (
	"context"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/product-catalog-service/internal/contract"
	"github.com/product-catalog-service/internal/domain"
	"github.com/product-catalog-service/internal/clock"
	"github.com/product-catalog-service/internal/committer"
)

// CreateProductRequest represents the input for creating a product.
type CreateProductRequest struct {
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
}

// CreateProductResponse represents the output of creating a product.
type CreateProductResponse struct {
	ProductID string
}

// UpdateProductRequest represents the input for updating a product.
type UpdateProductRequest struct {
	ProductID   string
	Name        string
	Description string
	Category    string
}

// ActivateProductRequest represents the input for activating a product.
type ActivateProductRequest struct {
	ProductID string
}

// DeactivateProductRequest represents the input for deactivating a product.
type DeactivateProductRequest struct {
	ProductID string
}

// ArchiveProductRequest represents the input for archiving a product.
type ArchiveProductRequest struct {
	ProductID string
}

// ApplyDiscountRequest represents the input for applying a discount to a product.
type ApplyDiscountRequest struct {
	ProductID          string
	DiscountPercentage float64
	StartDate          time.Time
	EndDate            time.Time
}

// RemoveDiscountRequest represents the input for removing a discount from a product.
type RemoveDiscountRequest struct {
	ProductID string
}

// ProductUseCases provides all product-related use cases.
type ProductUseCases struct {
	repo       contract.ProductRepository
	outboxRepo contract.OutboxRepository
	committer  *committer.Committer
	clock      clock.Clock
}

// NewProductUseCases creates a new ProductUseCases instance.
func NewProductUseCases(
	repo contract.ProductRepository,
	outboxRepo contract.OutboxRepository,
	committer *committer.Committer,
	clock clock.Clock,
) *ProductUseCases {
	return &ProductUseCases{
		repo:       repo,
		outboxRepo: outboxRepo,
		committer:  committer,
		clock:      clock,
	}
}

// CreateProduct creates a new product.
func (uc *ProductUseCases) CreateProduct(ctx context.Context, req CreateProductRequest) (*CreateProductResponse, error) {
	productID := uuid.New().String()
	basePrice := domain.NewMoney(req.BasePriceNumerator, req.BasePriceDenominator)
	now := uc.clock.Now()

	product, err := domain.NewProduct(
		productID,
		req.Name,
		req.Description,
		req.Category,
		basePrice,
		now,
	)
	if err != nil {
		return nil, err
	}

	plan := committer.NewPlan()

	if mut := uc.repo.InsertMut(product); mut != nil {
		plan.Add(mut)
	}

	for _, event := range product.DomainEvents() {
		if mut := uc.outboxRepo.InsertDomainEventMut(event); mut != nil {
			plan.Add(mut)
		}
	}

	if err := uc.committer.Apply(ctx, plan); err != nil {
		return nil, err
	}

	return &CreateProductResponse{ProductID: productID}, nil
}

// UpdateProduct updates an existing product.
func (uc *ProductUseCases) UpdateProduct(ctx context.Context, req UpdateProductRequest) error {
	product, err := uc.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	now := uc.clock.Now()
	if err := product.Update(req.Name, req.Description, req.Category, now); err != nil {
		return err
	}

	plan := committer.NewPlan()

	if mut := uc.repo.UpdateMut(product); mut != nil {
		plan.Add(mut)
	}

	for _, event := range product.DomainEvents() {
		if mut := uc.outboxRepo.InsertDomainEventMut(event); mut != nil {
			plan.Add(mut)
		}
	}

	if !plan.IsEmpty() {
		if err := uc.committer.Apply(ctx, plan); err != nil {
			return err
		}
	}

	return nil
}

// ActivateProduct activates a product.
func (uc *ProductUseCases) ActivateProduct(ctx context.Context, req ActivateProductRequest) error {
	product, err := uc.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	now := uc.clock.Now()
	if err := product.Activate(now); err != nil {
		return err
	}

	plan := committer.NewPlan()

	if mut := uc.repo.UpdateMut(product); mut != nil {
		plan.Add(mut)
	}

	for _, event := range product.DomainEvents() {
		if mut := uc.outboxRepo.InsertDomainEventMut(event); mut != nil {
			plan.Add(mut)
		}
	}

	if !plan.IsEmpty() {
		if err := uc.committer.Apply(ctx, plan); err != nil {
			return err
		}
	}

	return nil
}

// DeactivateProduct deactivates a product.
func (uc *ProductUseCases) DeactivateProduct(ctx context.Context, req DeactivateProductRequest) error {
	product, err := uc.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	now := uc.clock.Now()
	if err := product.Deactivate(now); err != nil {
		return err
	}

	plan := committer.NewPlan()

	if mut := uc.repo.UpdateMut(product); mut != nil {
		plan.Add(mut)
	}

	for _, event := range product.DomainEvents() {
		if mut := uc.outboxRepo.InsertDomainEventMut(event); mut != nil {
			plan.Add(mut)
		}
	}

	if !plan.IsEmpty() {
		if err := uc.committer.Apply(ctx, plan); err != nil {
			return err
		}
	}

	return nil
}

// ArchiveProduct archives a product (soft delete).
func (uc *ProductUseCases) ArchiveProduct(ctx context.Context, req ArchiveProductRequest) error {
	product, err := uc.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	now := uc.clock.Now()
	if err := product.Archive(now); err != nil {
		return err
	}

	plan := committer.NewPlan()

	if mut := uc.repo.ArchiveMut(product); mut != nil {
		plan.Add(mut)
	}

	for _, event := range product.DomainEvents() {
		if mut := uc.outboxRepo.InsertDomainEventMut(event); mut != nil {
			plan.Add(mut)
		}
	}

	if !plan.IsEmpty() {
		if err := uc.committer.Apply(ctx, plan); err != nil {
			return err
		}
	}

	return nil
}

// ApplyDiscount applies a discount to a product.
func (uc *ProductUseCases) ApplyDiscount(ctx context.Context, req ApplyDiscountRequest) error {
	product, err := uc.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	percentage := big.NewRat(int64(req.DiscountPercentage*100), 100)
	discount, err := domain.NewDiscount(percentage, req.StartDate, req.EndDate)
	if err != nil {
		return err
	}

	now := uc.clock.Now()
	if err := product.ApplyDiscount(discount, now); err != nil {
		return err
	}

	plan := committer.NewPlan()

	if mut := uc.repo.UpdateMut(product); mut != nil {
		plan.Add(mut)
	}

	for _, event := range product.DomainEvents() {
		if mut := uc.outboxRepo.InsertDomainEventMut(event); mut != nil {
			plan.Add(mut)
		}
	}

	if !plan.IsEmpty() {
		if err := uc.committer.Apply(ctx, plan); err != nil {
			return err
		}
	}

	return nil
}

// RemoveDiscount removes a discount from a product.
func (uc *ProductUseCases) RemoveDiscount(ctx context.Context, req RemoveDiscountRequest) error {
	product, err := uc.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	now := uc.clock.Now()
	if err := product.RemoveDiscount(now); err != nil {
		return err
	}

	plan := committer.NewPlan()

	if mut := uc.repo.UpdateMut(product); mut != nil {
		plan.Add(mut)
	}

	for _, event := range product.DomainEvents() {
		if mut := uc.outboxRepo.InsertDomainEventMut(event); mut != nil {
			plan.Add(mut)
		}
	}

	if !plan.IsEmpty() {
		if err := uc.committer.Apply(ctx, plan); err != nil {
			return err
		}
	}

	return nil
}

// ValidateCreateProductRequest validates the create product request.
func ValidateCreateProductRequest(req CreateProductRequest) error {
	if req.Name == "" {
		return domain.ErrInvalidProductName
	}
	if req.Category == "" {
		return domain.ErrInvalidProductCategory
	}
	if req.BasePriceNumerator <= 0 || req.BasePriceDenominator <= 0 {
		return domain.ErrInvalidBasePrice
	}
	price := big.NewRat(req.BasePriceNumerator, req.BasePriceDenominator)
	if price.Sign() <= 0 {
		return domain.ErrInvalidBasePrice
	}
	return nil
}

// ValidateUpdateProductRequest validates the update product request.
func ValidateUpdateProductRequest(req UpdateProductRequest) error {
	if req.ProductID == "" {
		return domain.ErrInvalidID
	}
	if req.Name == "" {
		return domain.ErrInvalidProductName
	}
	if req.Category == "" {
		return domain.ErrInvalidProductCategory
	}
	return nil
}

// ValidateProductIDRequest validates requests that require only a product ID.
func ValidateProductIDRequest(productID string) error {
	if productID == "" {
		return domain.ErrInvalidID
	}
	return nil
}

// ValidateApplyDiscountRequest validates the apply discount request.
func ValidateApplyDiscountRequest(req ApplyDiscountRequest) error {
	if req.ProductID == "" {
		return domain.ErrInvalidID
	}
	if req.DiscountPercentage <= 0 || req.DiscountPercentage > 100 {
		return domain.ErrInvalidDiscountPercentage
	}
	if !req.EndDate.After(req.StartDate) {
		return domain.ErrInvalidDiscountPeriod
	}
	return nil
}
