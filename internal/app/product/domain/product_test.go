package domain

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProduct_Valid(t *testing.T) {
	now := time.Now()
	basePrice := NewMoney(1999, 100)

	product, err := NewProduct("prod-123", "Test Product", "A description", "Electronics", basePrice, now)

	require.NoError(t, err)
	assert.Equal(t, "prod-123", product.ID())
	assert.Equal(t, "Test Product", product.Name())
	assert.Equal(t, "A description", product.Description())
	assert.Equal(t, "Electronics", product.Category())
	assert.Equal(t, ProductStatusDraft, product.Status())
	assert.NotNil(t, product.BasePrice())
	assert.Nil(t, product.Discount())
	assert.Len(t, product.DomainEvents(), 1)
	assert.IsType(t, ProductCreatedEvent{}, product.DomainEvents()[0])
}

func TestNewProduct_InvalidInputs(t *testing.T) {
	now := time.Now()
	basePrice := NewMoney(1999, 100)

	tests := []struct {
		name        string
		id          string
		productName string
		description string
		category    string
		price       *Money
		wantErr     error
	}{
		{
			name:        "empty id",
			id:          "",
			productName: "Test",
			category:    "Cat",
			price:       basePrice,
			wantErr:     ErrInvalidID,
		},
		{
			name:        "empty name",
			id:          "123",
			productName: "",
			category:    "Cat",
			price:       basePrice,
			wantErr:     ErrInvalidProductName,
		},
		{
			name:        "empty category",
			id:          "123",
			productName: "Test",
			category:    "",
			price:       basePrice,
			wantErr:     ErrInvalidProductCategory,
		},
		{
			name:        "nil price",
			id:          "123",
			productName: "Test",
			category:    "Cat",
			price:       nil,
			wantErr:     ErrInvalidBasePrice,
		},
		{
			name:        "zero price",
			id:          "123",
			productName: "Test",
			category:    "Cat",
			price:       Zero(),
			wantErr:     ErrInvalidBasePrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewProduct(tt.id, tt.productName, tt.description, tt.category, tt.price, now)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestProduct_Activate(t *testing.T) {
	now := time.Now()
	basePrice := NewMoney(1999, 100)
	product, _ := NewProduct("123", "Test", "Desc", "Cat", basePrice, now)

	product.ClearEvents() // Clear creation event

	err := product.Activate(now.Add(time.Hour))

	require.NoError(t, err)
	assert.Equal(t, ProductStatusActive, product.Status())
	assert.True(t, product.Changes().Dirty(FieldStatus))
	assert.Len(t, product.DomainEvents(), 1)
	assert.IsType(t, ProductActivatedEvent{}, product.DomainEvents()[0])
}

func TestProduct_Activate_AlreadyActive(t *testing.T) {
	now := time.Now()
	basePrice := NewMoney(1999, 100)
	product, _ := NewProduct("123", "Test", "Desc", "Cat", basePrice, now)
	product.Activate(now)

	err := product.Activate(now.Add(time.Hour))

	assert.ErrorIs(t, err, ErrProductAlreadyActive)
}

func TestProduct_Deactivate(t *testing.T) {
	now := time.Now()
	basePrice := NewMoney(1999, 100)
	product, _ := NewProduct("123", "Test", "Desc", "Cat", basePrice, now)
	product.Activate(now)
	product.ClearEvents()

	err := product.Deactivate(now.Add(time.Hour))

	require.NoError(t, err)
	assert.Equal(t, ProductStatusInactive, product.Status())
	assert.Len(t, product.DomainEvents(), 1)
	assert.IsType(t, ProductDeactivatedEvent{}, product.DomainEvents()[0])
}

func TestProduct_Archive(t *testing.T) {
	now := time.Now()
	basePrice := NewMoney(1999, 100)
	product, _ := NewProduct("123", "Test", "Desc", "Cat", basePrice, now)
	product.ClearEvents()

	err := product.Archive(now.Add(time.Hour))

	require.NoError(t, err)
	assert.Equal(t, ProductStatusArchived, product.Status())
	assert.NotNil(t, product.ArchivedAt())
	assert.Len(t, product.DomainEvents(), 1)
	assert.IsType(t, ProductArchivedEvent{}, product.DomainEvents()[0])
}

func TestProduct_ApplyDiscount(t *testing.T) {
	now := time.Now()
	basePrice := NewMoney(10000, 100) // $100.00
	product, _ := NewProduct("123", "Test", "Desc", "Cat", basePrice, now)
	product.Activate(now)
	product.ClearEvents()

	discount, _ := NewDiscount(big.NewRat(20, 1), now, now.Add(24*time.Hour))
	err := product.ApplyDiscount(discount, now)

	require.NoError(t, err)
	assert.NotNil(t, product.Discount())
	assert.True(t, product.HasActiveDiscount(now))
	assert.Len(t, product.DomainEvents(), 1)
	assert.IsType(t, DiscountAppliedEvent{}, product.DomainEvents()[0])

	// Check effective price
	effectivePrice := product.EffectivePrice(now)
	expected := NewMoney(8000, 100) // $80.00
	assert.True(t, effectivePrice.Equals(expected))
}

func TestProduct_ApplyDiscount_NotActive(t *testing.T) {
	now := time.Now()
	basePrice := NewMoney(10000, 100)
	product, _ := NewProduct("123", "Test", "Desc", "Cat", basePrice, now)
	// Product is in draft status

	discount, _ := NewDiscount(big.NewRat(20, 1), now, now.Add(24*time.Hour))
	err := product.ApplyDiscount(discount, now)

	assert.ErrorIs(t, err, ErrProductNotActive)
}

func TestProduct_RemoveDiscount(t *testing.T) {
	now := time.Now()
	basePrice := NewMoney(10000, 100)
	product, _ := NewProduct("123", "Test", "Desc", "Cat", basePrice, now)
	product.Activate(now)
	discount, _ := NewDiscount(big.NewRat(20, 1), now, now.Add(24*time.Hour))
	product.ApplyDiscount(discount, now)
	product.ClearEvents()

	err := product.RemoveDiscount(now.Add(time.Hour))

	require.NoError(t, err)
	assert.Nil(t, product.Discount())
	assert.False(t, product.HasActiveDiscount(now))
	assert.Len(t, product.DomainEvents(), 1)
	assert.IsType(t, DiscountRemovedEvent{}, product.DomainEvents()[0])
}

func TestProduct_RemoveDiscount_NoDiscount(t *testing.T) {
	now := time.Now()
	basePrice := NewMoney(10000, 100)
	product, _ := NewProduct("123", "Test", "Desc", "Cat", basePrice, now)

	err := product.RemoveDiscount(now)

	assert.ErrorIs(t, err, ErrNoDiscountToRemove)
}

func TestProduct_Update(t *testing.T) {
	now := time.Now()
	basePrice := NewMoney(1999, 100)
	product, _ := NewProduct("123", "Original", "Desc", "Cat", basePrice, now)
	product.ClearEvents()

	err := product.Update("Updated", "New Desc", "NewCat", now.Add(time.Hour))

	require.NoError(t, err)
	assert.Equal(t, "Updated", product.Name())
	assert.Equal(t, "New Desc", product.Description())
	assert.Equal(t, "NewCat", product.Category())
	assert.True(t, product.Changes().Dirty(FieldName))
	assert.True(t, product.Changes().Dirty(FieldDescription))
	assert.True(t, product.Changes().Dirty(FieldCategory))
	assert.Len(t, product.DomainEvents(), 1)
	assert.IsType(t, ProductUpdatedEvent{}, product.DomainEvents()[0])
}

func TestProduct_Update_Archived(t *testing.T) {
	now := time.Now()
	basePrice := NewMoney(1999, 100)
	product, _ := NewProduct("123", "Test", "Desc", "Cat", basePrice, now)
	product.Archive(now)

	err := product.Update("New", "Desc", "Cat", now.Add(time.Hour))

	assert.ErrorIs(t, err, ErrProductArchived)
}

func TestProduct_EffectivePrice_WithoutDiscount(t *testing.T) {
	now := time.Now()
	basePrice := NewMoney(5000, 100) // $50.00
	product, _ := NewProduct("123", "Test", "Desc", "Cat", basePrice, now)

	effectivePrice := product.EffectivePrice(now)

	assert.True(t, effectivePrice.Equals(basePrice))
}

func TestProduct_EffectivePrice_WithExpiredDiscount(t *testing.T) {
	now := time.Now()
	basePrice := NewMoney(10000, 100)
	product, _ := NewProduct("123", "Test", "Desc", "Cat", basePrice, now)
	product.Activate(now)

	// Apply a discount that ends before "now"
	discount, _ := NewDiscount(big.NewRat(20, 1), now.Add(-48*time.Hour), now.Add(-24*time.Hour))
	product.ApplyDiscount(discount, now.Add(-48*time.Hour))

	// Check effective price at current time (discount expired)
	effectivePrice := product.EffectivePrice(now)

	// Should be base price since discount expired
	assert.True(t, effectivePrice.Equals(basePrice))
}
