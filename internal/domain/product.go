package domain

import (
	"strings"
	"time"
)

// Product is the aggregate root for product management.
// It encapsulates all business logic related to products.
type Product struct {
	id          string
	name        string
	description string
	category    string
	basePrice   *Money
	discount    *Discount
	status      ProductStatus
	createdAt   time.Time
	updatedAt   time.Time
	archivedAt  *time.Time
	changes     *ChangeTracker
	events      []DomainEvent
}

// NewProduct creates a new Product aggregate.
func NewProduct(id, name, description, category string, basePrice *Money, now time.Time) (*Product, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidID
	}
	if strings.TrimSpace(name) == "" {
		return nil, ErrInvalidProductName
	}
	if strings.TrimSpace(category) == "" {
		return nil, ErrInvalidProductCategory
	}
	if basePrice == nil || !basePrice.IsPositive() {
		return nil, ErrInvalidBasePrice
	}

	p := &Product{
		id:          id,
		name:        strings.TrimSpace(name),
		description: strings.TrimSpace(description),
		category:    strings.TrimSpace(category),
		basePrice:   basePrice,
		status:      ProductStatusDraft,
		createdAt:   now,
		updatedAt:   now,
		changes:     NewChangeTracker(),
		events:      make([]DomainEvent, 0),
	}

	// Mark all fields as dirty for a new product
	p.changes.MarkAllDirty(FieldName, FieldDescription, FieldCategory, FieldBasePrice, FieldStatus)

	// Record the creation event
	p.events = append(p.events, NewProductCreatedEvent(
		id, p.name, p.description, p.category, p.basePrice, now,
	))

	return p, nil
}

// ReconstructProduct reconstructs a Product from persistence.
// This is used by repositories to load existing products.
func ReconstructProduct(
	id, name, description, category string,
	basePrice *Money,
	discount *Discount,
	status ProductStatus,
	createdAt, updatedAt time.Time,
	archivedAt *time.Time,
) *Product {
	return &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		discount:    discount,
		status:      status,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		archivedAt:  archivedAt,
		changes:     NewChangeTracker(),
		events:      make([]DomainEvent, 0),
	}
}

// Getters

func (p *Product) ID() string           { return p.id }
func (p *Product) Name() string         { return p.name }
func (p *Product) Description() string  { return p.description }
func (p *Product) Category() string     { return p.category }
func (p *Product) BasePrice() *Money    { return p.basePrice }
func (p *Product) Discount() *Discount  { return p.discount }
func (p *Product) Status() ProductStatus { return p.status }
func (p *Product) CreatedAt() time.Time { return p.createdAt }
func (p *Product) UpdatedAt() time.Time { return p.updatedAt }
func (p *Product) ArchivedAt() *time.Time { return p.archivedAt }
func (p *Product) Changes() *ChangeTracker { return p.changes }
func (p *Product) DomainEvents() []DomainEvent { return p.events }

// ClearEvents clears all domain events (typically after they've been processed).
func (p *Product) ClearEvents() {
	p.events = make([]DomainEvent, 0)
}

// EffectivePrice calculates the current effective price considering any active discount.
func (p *Product) EffectivePrice(now time.Time) *Money {
	if p.discount != nil && p.discount.IsActive(now) {
		return p.discount.ApplyTo(p.basePrice)
	}
	return p.basePrice
}

// HasActiveDiscount returns true if the product has an active discount at the given time.
func (p *Product) HasActiveDiscount(now time.Time) bool {
	return p.discount != nil && p.discount.IsActive(now)
}

// Business Methods

// Update updates the product details (name, description, category).
func (p *Product) Update(name, description, category string, now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}

	if strings.TrimSpace(name) == "" {
		return ErrInvalidProductName
	}
	if strings.TrimSpace(category) == "" {
		return ErrInvalidProductCategory
	}

	hasChanges := false

	newName := strings.TrimSpace(name)
	if p.name != newName {
		p.name = newName
		p.changes.MarkDirty(FieldName)
		hasChanges = true
	}

	newDescription := strings.TrimSpace(description)
	if p.description != newDescription {
		p.description = newDescription
		p.changes.MarkDirty(FieldDescription)
		hasChanges = true
	}

	newCategory := strings.TrimSpace(category)
	if p.category != newCategory {
		p.category = newCategory
		p.changes.MarkDirty(FieldCategory)
		hasChanges = true
	}

	if hasChanges {
		p.updatedAt = now
		p.events = append(p.events, NewProductUpdatedEvent(
			p.id, p.name, p.description, p.category, now,
		))
	}

	return nil
}

// Activate activates the product, making it available for sale.
func (p *Product) Activate(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}
	if p.status == ProductStatusActive {
		return ErrProductAlreadyActive
	}
	if !p.status.CanActivate() {
		return ErrProductNotActive
	}

	p.status = ProductStatusActive
	p.updatedAt = now
	p.changes.MarkDirty(FieldStatus)

	p.events = append(p.events, NewProductActivatedEvent(p.id, now))
	return nil
}

// Deactivate deactivates the product.
func (p *Product) Deactivate(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}
	if p.status == ProductStatusInactive {
		return ErrProductAlreadyInactive
	}
	if !p.status.CanDeactivate() {
		return ErrProductNotActive
	}

	p.status = ProductStatusInactive
	p.updatedAt = now
	p.changes.MarkDirty(FieldStatus)

	p.events = append(p.events, NewProductDeactivatedEvent(p.id, now))
	return nil
}

// Archive archives the product (soft delete).
func (p *Product) Archive(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}

	p.status = ProductStatusArchived
	p.archivedAt = &now
	p.updatedAt = now
	p.changes.MarkDirty(FieldStatus)

	p.events = append(p.events, NewProductArchivedEvent(p.id, now))
	return nil
}

// ApplyDiscount applies a discount to the product.
func (p *Product) ApplyDiscount(discount *Discount, now time.Time) error {
	if p.status != ProductStatusActive {
		return ErrProductNotActive
	}
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}

	if discount == nil {
		return ErrInvalidDiscountPercentage
	}

	// Validate discount is valid at the current time or starts in the future
	if discount.IsExpired(now) {
		return ErrInvalidDiscountPeriod
	}

	p.discount = discount
	p.updatedAt = now
	p.changes.MarkDirty(FieldDiscount)

	p.events = append(p.events, NewDiscountAppliedEvent(
		p.id, discount.Percentage(), discount.StartDate(), discount.EndDate(), now,
	))
	return nil
}

// RemoveDiscount removes the current discount from the product.
func (p *Product) RemoveDiscount(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}
	if p.discount == nil {
		return ErrNoDiscountToRemove
	}

	p.discount = nil
	p.updatedAt = now
	p.changes.MarkDirty(FieldDiscount)

	p.events = append(p.events, NewDiscountRemovedEvent(p.id, now))
	return nil
}

// IsActive returns true if the product is active.
func (p *Product) IsActive() bool {
	return p.status == ProductStatusActive
}

// IsArchived returns true if the product is archived.
func (p *Product) IsArchived() bool {
	return p.status == ProductStatusArchived
}
