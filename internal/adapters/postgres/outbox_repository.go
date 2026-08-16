package postgres

import (
	"context"
	"encoding/json"

	"br.com.nevvesdev/realtime-payment/internal/domain/errors"
	"br.com.nevvesdev/realtime-payment/internal/domain/repositories"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresOutboxRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresOutboxRepository(pool *pgxpool.Pool) repositories.OutboxRepository {
	return &PostgresOutboxRepository{pool: pool}
}

func (r *PostgresOutboxRepository) Save(ctx context.Context, event *repositories.OutboxEvent) error {
	query := `
		INSERT INTO outbox_events (aggregate_id, event_type, event_data, processed)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query,
		event.AggregateID,
		event.EventType,
		event.EventData,
		event.Processed,
	).Scan(&event.ID)

	if err != nil {
		return errors.NewDomainErrorWithCause(
			errors.ErrDatabaseError,
			"erro ao salvar evento na outbox",
			err,
		)
	}

	return nil
}

func (r *PostgresOutboxRepository) FindUnprocessed(ctx context.Context, limit int) ([]*repositories.OutboxEvent, error) {
	query := `
		SELECT id, aggregate_id, event_type, event_data, processed
		FROM outbox_events
		WHERE processed = false
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, errors.NewDomainErrorWithCause(
			errors.ErrDatabaseError,
			"erro ao buscar eventos não processados da outbox",
			err,
		)
	}
	defer rows.Close()

	var events []*repositories.OutboxEvent
	for rows.Next() {
		var (
			id          int64
			aggregateID uuid.UUID
			eventType   string
			eventData   json.RawMessage
			processed   bool
		)

		err := rows.Scan(&id, &aggregateID, &eventType, &eventData, &processed)
		if err != nil {
			return nil, errors.NewDomainErrorWithCause(
				errors.ErrDatabaseError,
				"erro ao processar linha de evento outbox",
				err,
			)
		}

		events = append(events, &repositories.OutboxEvent{
			ID:          id,
			AggregateID: aggregateID,
			EventType:   eventType,
			EventData:   eventData,
			Processed:   processed,
		})
	}

	return events, nil
}

func (r *PostgresOutboxRepository) MarkAsProcessed(ctx context.Context, id int64) error {
	query := `
		UPDATE outbox_events
		SET processed = true, processed_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return errors.NewDomainErrorWithCause(
			errors.ErrDatabaseError,
			"erro ao marcar evento outbox como processado",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return errors.NewDomainError(
			errors.ErrDatabaseError,
			"evento outbox não encontrado para atualização",
		)
	}

	return nil
}

func (r *PostgresOutboxRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM outbox_events WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return errors.NewDomainErrorWithCause(
			errors.ErrDatabaseError,
			"erro ao deletar evento outbox",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return errors.NewDomainError(
			errors.ErrDatabaseError,
			"evento outbox não encontrado para deleção",
		)
	}

	return nil
}
