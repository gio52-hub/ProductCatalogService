package contract

import (
	"cloud.google.com/go/spanner"
	"github.com/product-catalog-service/internal/domain"
)

// OutboxEvent represents an enriched event ready for persistence.
type OutboxEvent struct {
	EventID     string
	EventType   string
	AggregateID string
	Payload     interface{}
}

// OutboxRepository defines the interface for outbox event persistence.
type OutboxRepository interface {
	// InsertMut returns a mutation for inserting an outbox event.
	InsertMut(event *OutboxEvent) *spanner.Mutation

	// InsertDomainEventMut converts a domain event to an outbox event and returns a mutation.
	InsertDomainEventMut(event domain.DomainEvent) *spanner.Mutation
}
