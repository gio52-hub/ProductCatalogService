package domain

// ProductStatus represents the status of a product.
type ProductStatus string

// Product status values.
const (
	ProductStatusDraft    ProductStatus = "draft"
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
	ProductStatusArchived ProductStatus = "archived"
)

// String returns the string representation of the status.
func (s ProductStatus) String() string {
	return string(s)
}

// IsValid checks if the status is a valid product status.
func (s ProductStatus) IsValid() bool {
	switch s {
	case ProductStatusDraft, ProductStatusActive, ProductStatusInactive, ProductStatusArchived:
		return true
	default:
		return false
	}
}

// CanActivate returns true if a product with this status can be activated.
func (s ProductStatus) CanActivate() bool {
	return s == ProductStatusDraft || s == ProductStatusInactive
}

// CanDeactivate returns true if a product with this status can be deactivated.
func (s ProductStatus) CanDeactivate() bool {
	return s == ProductStatusActive
}

// CanArchive returns true if a product with this status can be archived.
func (s ProductStatus) CanArchive() bool {
	return s != ProductStatusArchived
}

// CanUpdate returns true if a product with this status can be updated.
func (s ProductStatus) CanUpdate() bool {
	return s != ProductStatusArchived
}

// CanApplyDiscount returns true if a discount can be applied to a product with this status.
func (s ProductStatus) CanApplyDiscount() bool {
	return s == ProductStatusActive
}
