// Package domain contains the core business logic, aggregates, value objects, and domain events.
package domain

// Field constants for change tracking
const (
	FieldName        = "name"
	FieldDescription = "description"
	FieldCategory    = "category"
	FieldBasePrice   = "base_price"
	FieldDiscount    = "discount"
	FieldStatus      = "status"
)

// ChangeTracker tracks which fields have been modified on an aggregate.
type ChangeTracker struct {
	dirtyFields map[string]bool
}

// NewChangeTracker creates a new ChangeTracker instance.
func NewChangeTracker() *ChangeTracker {
	return &ChangeTracker{
		dirtyFields: make(map[string]bool),
	}
}

// MarkDirty marks a field as dirty (modified).
func (ct *ChangeTracker) MarkDirty(field string) {
	if ct == nil {
		return
	}
	if ct.dirtyFields == nil {
		ct.dirtyFields = make(map[string]bool)
	}
	ct.dirtyFields[field] = true
}

// Dirty returns true if the specified field has been modified.
func (ct *ChangeTracker) Dirty(field string) bool {
	if ct == nil || ct.dirtyFields == nil {
		return false
	}
	return ct.dirtyFields[field]
}

// DirtyFields returns a slice of all dirty field names.
func (ct *ChangeTracker) DirtyFields() []string {
	if ct == nil || ct.dirtyFields == nil {
		return nil
	}
	fields := make([]string, 0, len(ct.dirtyFields))
	for field, dirty := range ct.dirtyFields {
		if dirty {
			fields = append(fields, field)
		}
	}
	return fields
}

// HasChanges returns true if any field has been modified.
func (ct *ChangeTracker) HasChanges() bool {
	if ct == nil || ct.dirtyFields == nil {
		return false
	}
	for _, dirty := range ct.dirtyFields {
		if dirty {
			return true
		}
	}
	return false
}

// Reset clears all dirty flags.
func (ct *ChangeTracker) Reset() {
	if ct == nil {
		return
	}
	ct.dirtyFields = make(map[string]bool)
}

// Clear removes a specific field from the dirty set.
func (ct *ChangeTracker) Clear(field string) {
	if ct == nil || ct.dirtyFields == nil {
		return
	}
	delete(ct.dirtyFields, field)
}

// MarkAllDirty marks all provided fields as dirty.
func (ct *ChangeTracker) MarkAllDirty(fields ...string) {
	for _, field := range fields {
		ct.MarkDirty(field)
	}
}
