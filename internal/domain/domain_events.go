package domain

import (
	"math/big"
	"time"
)

// DomainEvent is the interface that all domain events must implement.
//
//nolint:revive // stuttering is intentional for DDD ubiquitous language clarity
type DomainEvent interface {
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
}

// BaseEvent contains common fields for all domain events.
type BaseEvent struct {
	aggregateID string
	occurredAt  time.Time
}

// AggregateID returns the ID of the aggregate that raised the event.
func (e BaseEvent) AggregateID() string {
	return e.aggregateID
}

// OccurredAt returns the time the event occurred.
func (e BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// ProductCreatedEvent is raised when a new product is created.
type ProductCreatedEvent struct {
	BaseEvent
	Name        string
	Description string
	Category    string
	BasePrice   *Money
}

// EventType returns the event type identifier.
func (e ProductCreatedEvent) EventType() string {
	return "product.created"
}

// NewProductCreatedEvent creates a new ProductCreatedEvent.
func NewProductCreatedEvent(productID, name, description, category string, basePrice *Money, occurredAt time.Time) ProductCreatedEvent {
	return ProductCreatedEvent{
		BaseEvent: BaseEvent{
			aggregateID: productID,
			occurredAt:  occurredAt,
		},
		Name:        name,
		Description: description,
		Category:    category,
		BasePrice:   basePrice,
	}
}

// ProductUpdatedEvent is raised when product details are updated.
type ProductUpdatedEvent struct {
	BaseEvent
	Name        string
	Description string
	Category    string
}

// EventType returns the event type identifier.
func (e ProductUpdatedEvent) EventType() string {
	return "product.updated"
}

// NewProductUpdatedEvent creates a new ProductUpdatedEvent.
func NewProductUpdatedEvent(productID, name, description, category string, occurredAt time.Time) ProductUpdatedEvent {
	return ProductUpdatedEvent{
		BaseEvent: BaseEvent{
			aggregateID: productID,
			occurredAt:  occurredAt,
		},
		Name:        name,
		Description: description,
		Category:    category,
	}
}

// ProductActivatedEvent is raised when a product is activated.
type ProductActivatedEvent struct {
	BaseEvent
}

// EventType returns the event type identifier.
func (e ProductActivatedEvent) EventType() string {
	return "product.activated"
}

// NewProductActivatedEvent creates a new ProductActivatedEvent.
func NewProductActivatedEvent(productID string, occurredAt time.Time) ProductActivatedEvent {
	return ProductActivatedEvent{
		BaseEvent: BaseEvent{
			aggregateID: productID,
			occurredAt:  occurredAt,
		},
	}
}

// ProductDeactivatedEvent is raised when a product is deactivated.
type ProductDeactivatedEvent struct {
	BaseEvent
}

// EventType returns the event type identifier.
func (e ProductDeactivatedEvent) EventType() string {
	return "product.deactivated"
}

// NewProductDeactivatedEvent creates a new ProductDeactivatedEvent.
func NewProductDeactivatedEvent(productID string, occurredAt time.Time) ProductDeactivatedEvent {
	return ProductDeactivatedEvent{
		BaseEvent: BaseEvent{
			aggregateID: productID,
			occurredAt:  occurredAt,
		},
	}
}

// ProductArchivedEvent is raised when a product is archived (soft deleted).
type ProductArchivedEvent struct {
	BaseEvent
}

// EventType returns the event type identifier.
func (e ProductArchivedEvent) EventType() string {
	return "product.archived"
}

// NewProductArchivedEvent creates a new ProductArchivedEvent.
func NewProductArchivedEvent(productID string, occurredAt time.Time) ProductArchivedEvent {
	return ProductArchivedEvent{
		BaseEvent: BaseEvent{
			aggregateID: productID,
			occurredAt:  occurredAt,
		},
	}
}

// DiscountAppliedEvent is raised when a discount is applied to a product.
type DiscountAppliedEvent struct {
	BaseEvent
	DiscountPercentage *big.Rat
	StartDate          time.Time
	EndDate            time.Time
}

// EventType returns the event type identifier.
func (e DiscountAppliedEvent) EventType() string {
	return "product.discount_applied"
}

// NewDiscountAppliedEvent creates a new DiscountAppliedEvent.
func NewDiscountAppliedEvent(productID string, percentage *big.Rat, startDate, endDate, occurredAt time.Time) DiscountAppliedEvent {
	return DiscountAppliedEvent{
		BaseEvent: BaseEvent{
			aggregateID: productID,
			occurredAt:  occurredAt,
		},
		DiscountPercentage: percentage,
		StartDate:          startDate,
		EndDate:            endDate,
	}
}

// DiscountRemovedEvent is raised when a discount is removed from a product.
type DiscountRemovedEvent struct {
	BaseEvent
}

// EventType returns the event type identifier.
func (e DiscountRemovedEvent) EventType() string {
	return "product.discount_removed"
}

// NewDiscountRemovedEvent creates a new DiscountRemovedEvent.
func NewDiscountRemovedEvent(productID string, occurredAt time.Time) DiscountRemovedEvent {
	return DiscountRemovedEvent{
		BaseEvent: BaseEvent{
			aggregateID: productID,
			occurredAt:  occurredAt,
		},
	}
}
