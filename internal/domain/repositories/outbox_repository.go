package repositories

import (
	"context"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID          int64
	AggregateID uuid.UUID
	EventType   string
	EventData   []byte
	Processed   bool
}

type OutboxRepository interface {
	// Salvar evento na outbox
	Save(ctx context.Context, event *OutboxEvent) error

	// Buscar eventos não processados
	FindUnprocessed(ctx context.Context, limit int) ([]*OutboxEvent, error)

	// Marcar evento como processado
	MarkAsProcessed(ctx context.Context, id int64) error

	// Deletar evento processado
	Delete(ctx context.Context, id int64) error
}
