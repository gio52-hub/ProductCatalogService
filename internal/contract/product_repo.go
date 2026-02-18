package contract

import (
	"context"

	"cloud.google.com/go/spanner"
	"github.com/product-catalog-service/internal/domain"
)

// ProductRepository defines the interface for product persistence operations.
// Following the pattern where repositories return mutations instead of applying them.
type ProductRepository interface {
	// FindByID retrieves a product by its ID.
	FindByID(ctx context.Context, id string) (*domain.Product, error)

	// InsertMut returns a mutation for inserting a new product.
	// The mutation should be added to a Plan and applied by the use case.
	InsertMut(product *domain.Product) *spanner.Mutation

	// UpdateMut returns a mutation for updating an existing product.
	// Only changed fields (tracked by ChangeTracker) are included.
	// Returns nil if there are no changes.
	UpdateMut(product *domain.Product) *spanner.Mutation

	// ArchiveMut returns a mutation for archiving a product.
	ArchiveMut(product *domain.Product) *spanner.Mutation
}
