package m_outbox

import (
	"time"
)

const (
	StatusPending   = "pending"
	StatusProcessed = "processed"
)

// OutboxEvent represents a database row in the outbox_events table
type OutboxEvent struct {
	EventID     string
	EventType   string
	AggregateID string
	Payload     string // JSON payload
	Status      string
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

// ToMap converts the outbox event to a map for Spanner mutation
func (e *OutboxEvent) ToMap() map[string]interface{} {
	return map[string]interface{}{
		EventID:     e.EventID,
		EventType:   e.EventType,
		AggregateID: e.AggregateID,
		Payload:     e.Payload,
		Status:      e.Status,
		CreatedAt:   e.CreatedAt,
		ProcessedAt: e.ProcessedAt,
	}
}
