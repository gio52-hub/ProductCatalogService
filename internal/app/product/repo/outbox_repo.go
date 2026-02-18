package repo

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"github.com/product-catalog-service/internal/app/product/contracts"
	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/product-catalog-service/internal/models"
)

// OutboxRepo implements the OutboxRepository interface using Spanner.
type OutboxRepo struct {
	model *models.OutboxModel
}

// NewOutboxRepo creates a new OutboxRepo.
func NewOutboxRepo() *OutboxRepo {
	return &OutboxRepo{
		model: models.NewOutboxModel(),
	}
}

// InsertMut returns a mutation for inserting an outbox event.
func (r *OutboxRepo) InsertMut(event *contracts.OutboxEvent) *spanner.Mutation {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		payload = []byte("{}")
	}

	data := &models.OutboxEventData{
		EventID:     event.EventID,
		EventType:   event.EventType,
		AggregateID: event.AggregateID,
		Payload:     spanner.NullJSON{Value: json.RawMessage(payload), Valid: true},
		Status:      models.StatusPending,
		CreatedAt:   time.Now(),
	}

	return r.model.InsertMut(data)
}

// InsertDomainEventMut converts a domain event to an outbox event and returns a mutation.
func (r *OutboxRepo) InsertDomainEventMut(event domain.DomainEvent) *spanner.Mutation {
	outboxEvent := &contracts.OutboxEvent{
		EventID:     uuid.New().String(),
		EventType:   event.EventType(),
		AggregateID: event.AggregateID(),
		Payload:     r.domainEventToPayload(event),
	}
	return r.InsertMut(outboxEvent)
}

// domainEventToPayload converts a domain event to a JSON-serializable payload.
func (r *OutboxRepo) domainEventToPayload(event domain.DomainEvent) map[string]interface{} {
	payload := map[string]interface{}{
		"event_type":   event.EventType(),
		"aggregate_id": event.AggregateID(),
		"occurred_at":  event.OccurredAt(),
	}

	switch e := event.(type) {
	case domain.ProductCreatedEvent:
		payload["name"] = e.Name
		payload["description"] = e.Description
		payload["category"] = e.Category
		if e.BasePrice != nil {
			payload["base_price_numerator"] = e.BasePrice.Numerator()
			payload["base_price_denominator"] = e.BasePrice.Denominator()
		}

	case domain.ProductUpdatedEvent:
		payload["name"] = e.Name
		payload["description"] = e.Description
		payload["category"] = e.Category

	case domain.DiscountAppliedEvent:
		if e.DiscountPercentage != nil {
			f, _ := e.DiscountPercentage.Float64()
			payload["discount_percentage"] = f
		}
		payload["start_date"] = e.StartDate
		payload["end_date"] = e.EndDate

	case domain.ProductActivatedEvent:
		// No additional fields

	case domain.ProductDeactivatedEvent:
		// No additional fields

	case domain.ProductArchivedEvent:
		// No additional fields

	case domain.DiscountRemovedEvent:
		// No additional fields
	}

	return payload
}
