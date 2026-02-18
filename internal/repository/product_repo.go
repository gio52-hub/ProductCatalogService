package repository

import (
	"context"
	"math/big"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/product-catalog-service/internal/domain"
)

// ProductRepo implements the ProductRepository interface using Spanner.
type ProductRepo struct {
	client *spanner.Client
	model  *ProductModel
}

// NewProductRepo creates a new ProductRepo.
func NewProductRepo(client *spanner.Client) *ProductRepo {
	return &ProductRepo{
		client: client,
		model:  NewProductModel(),
	}
}

// FindByID retrieves a product by its ID.
func (r *ProductRepo) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	row, err := r.client.Single().ReadRow(
		ctx,
ProductsTable,
		spanner.Key{id},
ProductAllColumns(),
	)
	if err != nil {
		if spanner.ErrCode(err) == 5 { // NOT_FOUND
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}

	return r.rowToProduct(row)
}

// InsertMut returns a mutation for inserting a new product.
func (r *ProductRepo) InsertMut(product *domain.Product) *spanner.Mutation {
	data := r.productToData(product)
	return r.model.InsertMut(data)
}

// UpdateMut returns a mutation for updating an existing product.
// Only changed fields (tracked by ChangeTracker) are included.
func (r *ProductRepo) UpdateMut(product *domain.Product) *spanner.Mutation {
	changes := product.Changes()
	if !changes.HasChanges() {
		return nil
	}

	updates := make(map[string]interface{})

	if changes.Dirty(domain.FieldName) {
		updates[ProductName] = product.Name()
	}

	if changes.Dirty(domain.FieldDescription) {
		updates[ProductDescription] = product.Description()
	}

	if changes.Dirty(domain.FieldCategory) {
		updates[ProductCategory] = product.Category()
	}

	if changes.Dirty(domain.FieldBasePrice) {
		updates[ProductBasePriceNum] = product.BasePrice().Numerator()
		updates[ProductBasePriceDenom] = product.BasePrice().Denominator()
	}

	if changes.Dirty(domain.FieldDiscount) {
		discount := product.Discount()
		if discount != nil {
			pct, _ := discount.Percentage().Float64()
			updates[ProductDiscountPercent] = spanner.NullNumeric{
				Numeric: *big.NewRat(int64(pct*100), 100),
				Valid:   true,
			}
			updates[ProductDiscountStartDate] = spanner.NullTime{Time: discount.StartDate(), Valid: true}
			updates[ProductDiscountEndDate] = spanner.NullTime{Time: discount.EndDate(), Valid: true}
		} else {
			updates[ProductDiscountPercent] = spanner.NullNumeric{Valid: false}
			updates[ProductDiscountStartDate] = spanner.NullTime{Valid: false}
			updates[ProductDiscountEndDate] = spanner.NullTime{Valid: false}
		}
	}

	if changes.Dirty(domain.FieldStatus) {
		updates[ProductStatus] = product.Status().String()
		if product.IsArchived() && product.ArchivedAt() != nil {
			updates[ProductArchivedAt] = spanner.NullTime{Time: *product.ArchivedAt(), Valid: true}
		}
	}

	if len(updates) == 0 {
		return nil
	}

	updates[ProductUpdatedAt] = product.UpdatedAt()
	return r.model.UpdateMut(product.ID(), updates)
}

// ArchiveMut returns a mutation for archiving a product.
func (r *ProductRepo) ArchiveMut(product *domain.Product) *spanner.Mutation {
	updates := map[string]interface{}{
ProductStatus:    product.Status().String(),
ProductUpdatedAt: product.UpdatedAt(),
	}
	if product.ArchivedAt() != nil {
		updates[ProductArchivedAt] = spanner.NullTime{Time: *product.ArchivedAt(), Valid: true}
	}
	return r.model.UpdateMut(product.ID(), updates)
}

// productToData converts a domain Product to a database model.
func (r *ProductRepo) productToData(product *domain.Product) *ProductData {
	data := &ProductData{
		ProductID:            product.ID(),
		Name:                 product.Name(),
		Description:          product.Description(),
		Category:             product.Category(),
		BasePriceNumerator:   product.BasePrice().Numerator(),
		BasePriceDenominator: product.BasePrice().Denominator(),
		Status:               product.Status().String(),
		CreatedAt:            product.CreatedAt(),
		UpdatedAt:            product.UpdatedAt(),
	}

	if discount := product.Discount(); discount != nil {
		pct, _ := discount.Percentage().Float64()
		data.DiscountPercent = spanner.NullNumeric{
			Numeric: *big.NewRat(int64(pct*100), 100),
			Valid:   true,
		}
		data.DiscountStartDate = spanner.NullTime{Time: discount.StartDate(), Valid: true}
		data.DiscountEndDate = spanner.NullTime{Time: discount.EndDate(), Valid: true}
	}

	if archivedAt := product.ArchivedAt(); archivedAt != nil {
		data.ArchivedAt = spanner.NullTime{Time: *archivedAt, Valid: true}
	}

	return data
}

// rowToProduct converts a Spanner row to a domain Product.
func (r *ProductRepo) rowToProduct(row *spanner.Row) (*domain.Product, error) {
	var data ProductData

	if err := row.Columns(
		&data.ProductID,
		&data.Name,
		&data.Description,
		&data.Category,
		&data.BasePriceNumerator,
		&data.BasePriceDenominator,
		&data.DiscountPercent,
		&data.DiscountStartDate,
		&data.DiscountEndDate,
		&data.Status,
		&data.CreatedAt,
		&data.UpdatedAt,
		&data.ArchivedAt,
	); err != nil {
		return nil, err
	}

	return r.dataToDomain(&data)
}

// dataToDomain converts a database model to a domain Product.
func (r *ProductRepo) dataToDomain(data *ProductData) (*domain.Product, error) {
	basePrice := domain.NewMoney(data.BasePriceNumerator, data.BasePriceDenominator)

	var discount *domain.Discount
	if data.DiscountPercent.Valid && data.DiscountStartDate.Valid && data.DiscountEndDate.Valid {
		pct, _ := data.DiscountPercent.Numeric.Float64()
		var err error
		discount, err = domain.NewDiscount(
			big.NewRat(int64(pct), 1),
			data.DiscountStartDate.Time,
			data.DiscountEndDate.Time,
		)
		if err != nil {
			// If discount is invalid, ignore it
			discount = nil
		}
	}

	var archivedAt *time.Time
	if data.ArchivedAt.Valid {
		archivedAt = &data.ArchivedAt.Time
	}

	return domain.ReconstructProduct(
		data.ProductID,
		data.Name,
		data.Description,
		data.Category,
		basePrice,
		discount,
		domain.ProductStatus(data.Status),
		data.CreatedAt,
		data.UpdatedAt,
		archivedAt,
	), nil
}
