// Package repository implements the persistence layer using Google Cloud Spanner.
package repository

import (
	"time"

	"cloud.google.com/go/spanner"
)

// Product table constants
const (
	ProductsTable            = "products"
	ProductID                = "product_id"
	ProductName              = "name"
	ProductDescription       = "description"
	ProductCategory          = "category"
	ProductBasePriceNum      = "base_price_numerator"
	ProductBasePriceDenom    = "base_price_denominator"
	ProductDiscountPercent   = "discount_percent"
	ProductDiscountStartDate = "discount_start_date"
	ProductDiscountEndDate   = "discount_end_date"
	ProductStatus            = "status"
	ProductCreatedAt         = "created_at"
	ProductUpdatedAt         = "updated_at"
	ProductArchivedAt        = "archived_at"
)

// Outbox table constants
const (
	OutboxTable       = "outbox_events"
	OutboxEventID     = "event_id"
	OutboxEventType   = "event_type"
	OutboxAggregateID = "aggregate_id"
	OutboxPayload     = "payload"
	OutboxStatus      = "status"
	OutboxCreatedAt   = "created_at"
	OutboxProcessedAt = "processed_at"
)

// Outbox event status constants
const (
	StatusPending   = "pending"
	StatusProcessed = "processed"
	StatusFailed    = "failed"
)

// ProductData represents the database model for a product.
type ProductData struct {
	ProductID            string
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
	DiscountPercent      spanner.NullNumeric
	DiscountStartDate    spanner.NullTime
	DiscountEndDate      spanner.NullTime
	Status               string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ArchivedAt           spanner.NullTime
}

// InsertMap returns a map of column names to values for INSERT operations.
func (p *ProductData) InsertMap() map[string]interface{} {
	return map[string]interface{}{
		ProductID:                p.ProductID,
		ProductName:              p.Name,
		ProductDescription:       p.Description,
		ProductCategory:          p.Category,
		ProductBasePriceNum:      p.BasePriceNumerator,
		ProductBasePriceDenom:    p.BasePriceDenominator,
		ProductDiscountPercent:   p.DiscountPercent,
		ProductDiscountStartDate: p.DiscountStartDate,
		ProductDiscountEndDate:   p.DiscountEndDate,
		ProductStatus:            p.Status,
		ProductCreatedAt:         p.CreatedAt,
		ProductUpdatedAt:         p.UpdatedAt,
		ProductArchivedAt:        p.ArchivedAt,
	}
}

// InsertMutation creates a Spanner mutation for inserting a product.
func (p *ProductData) InsertMutation() *spanner.Mutation {
	return spanner.InsertMap(ProductsTable, p.InsertMap())
}

// ProductAllColumns returns all column names for the products table.
func ProductAllColumns() []string {
	return []string{
		ProductID,
		ProductName,
		ProductDescription,
		ProductCategory,
		ProductBasePriceNum,
		ProductBasePriceDenom,
		ProductDiscountPercent,
		ProductDiscountStartDate,
		ProductDiscountEndDate,
		ProductStatus,
		ProductCreatedAt,
		ProductUpdatedAt,
		ProductArchivedAt,
	}
}

// OutboxEventData represents the database model for an outbox event.
type OutboxEventData struct {
	EventID     string
	EventType   string
	AggregateID string
	Payload     spanner.NullJSON
	Status      string
	CreatedAt   time.Time
	ProcessedAt spanner.NullTime
}

// InsertMap returns a map of column names to values for INSERT operations.
func (e *OutboxEventData) InsertMap() map[string]interface{} {
	return map[string]interface{}{
		OutboxEventID:     e.EventID,
		OutboxEventType:   e.EventType,
		OutboxAggregateID: e.AggregateID,
		OutboxPayload:     e.Payload,
		OutboxStatus:      e.Status,
		OutboxCreatedAt:   e.CreatedAt,
		OutboxProcessedAt: e.ProcessedAt,
	}
}

// InsertMutation creates a Spanner mutation for inserting an outbox event.
func (e *OutboxEventData) InsertMutation() *spanner.Mutation {
	return spanner.InsertMap(OutboxTable, e.InsertMap())
}

// OutboxAllColumns returns all column names for the outbox_events table.
func OutboxAllColumns() []string {
	return []string{
		OutboxEventID,
		OutboxEventType,
		OutboxAggregateID,
		OutboxPayload,
		OutboxStatus,
		OutboxCreatedAt,
		OutboxProcessedAt,
	}
}

// ProductModel provides helper methods for building product Spanner mutations.
type ProductModel struct{}

// NewProductModel creates a new ProductModel instance.
func NewProductModel() *ProductModel {
	return &ProductModel{}
}

// InsertMut creates an INSERT mutation from ProductData.
func (m *ProductModel) InsertMut(data *ProductData) *spanner.Mutation {
	return data.InsertMutation()
}

// UpdateMut creates an UPDATE mutation with the given updates.
func (m *ProductModel) UpdateMut(productID string, updates map[string]interface{}) *spanner.Mutation {
	if len(updates) == 0 {
		return nil
	}
	updates[ProductID] = productID
	return spanner.UpdateMap(ProductsTable, updates)
}

// OutboxModel provides helper methods for building outbox Spanner mutations.
type OutboxModel struct{}

// NewOutboxModel creates a new OutboxModel instance.
func NewOutboxModel() *OutboxModel {
	return &OutboxModel{}
}

// InsertMut creates an INSERT mutation from OutboxEventData.
func (m *OutboxModel) InsertMut(data *OutboxEventData) *spanner.Mutation {
	return data.InsertMutation()
}

// UpdateMut creates an UPDATE mutation with the given updates.
func (m *OutboxModel) UpdateMut(eventID string, updates map[string]interface{}) *spanner.Mutation {
	if len(updates) == 0 {
		return nil
	}
	updates[OutboxEventID] = eventID
	return spanner.UpdateMap(OutboxTable, updates)
}
