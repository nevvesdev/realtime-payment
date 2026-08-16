package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"br.com.nevvesdev/realtime-payment/internal/domain/entities"
	"br.com.nevvesdev/realtime-payment/internal/domain/errors"
	"br.com.nevvesdev/realtime-payment/internal/domain/repositories"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPaymentEventRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPaymentEventRepository(pool *pgxpool.Pool) repositories.PaymentEventRepository {
	return &PostgresPaymentEventRepository{pool: pool}
}

func (r *PostgresPaymentEventRepository) Save(ctx context.Context, event *entities.PaymentEvent) error {
	query := `
		INSERT INTO payment_events (payment_id, event_type, event_data, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query,
		event.PaymentID,
		event.EventType,
		event.EventData,
		event.CreatedAt,
	)

	if err != nil {
		return errors.NewDomainErrorWithCause(
			errors.ErrDatabaseError,
			"erro ao salvar evento de pagamento",
			err,
		)
	}

	return nil
}

func (r *PostgresPaymentEventRepository) FindByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*entities.PaymentEvent, error) {
	query := `
		SELECT id, payment_id, event_type, event_data, created_at
		FROM payment_events
		WHERE payment_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, paymentID)
	if err != nil {
		return nil, errors.NewDomainErrorWithCause(
			errors.ErrDatabaseError,
			"erro ao buscar eventos de pagamento",
			err,
		)
	}
	defer rows.Close()

	var events []*entities.PaymentEvent
	for rows.Next() {
		event, err := r.scanPaymentEventFromRows(rows)
		if err != nil {
			return nil, errors.NewDomainErrorWithCause(
				errors.ErrDatabaseError,
				"erro ao processar linha de evento",
				err,
			)
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *PostgresPaymentEventRepository) FindByPaymentIDWithPagination(
	ctx context.Context,
	paymentID uuid.UUID,
	limit, offset int,
) ([]*entities.PaymentEvent, error) {
	query := `
		SELECT id, payment_id, event_type, event_data, created_at
		FROM payment_events
		WHERE payment_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, paymentID, limit, offset)
	if err != nil {
		return nil, errors.NewDomainErrorWithCause(
			errors.ErrDatabaseError,
			"erro ao buscar eventos de pagamento com paginação",
			err,
		)
	}
	defer rows.Close()

	var events []*entities.PaymentEvent
	for rows.Next() {
		event, err := r.scanPaymentEventFromRows(rows)
		if err != nil {
			return nil, errors.NewDomainErrorWithCause(
				errors.ErrDatabaseError,
				"erro ao processar linha de evento",
				err,
			)
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *PostgresPaymentEventRepository) CountByPaymentID(ctx context.Context, paymentID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM payment_events WHERE payment_id = $1`

	err := r.pool.QueryRow(ctx, query, paymentID).Scan(&count)
	if err != nil {
		return 0, errors.NewDomainErrorWithCause(
			errors.ErrDatabaseError,
			"erro ao contar eventos de pagamento",
			err,
		)
	}

	return count, nil
}

func (r *PostgresPaymentEventRepository) scanPaymentEventFromRows(rows interface {
	Scan(dest ...interface{}) error
}) (*entities.PaymentEvent, error) {
	var (
		id        uuid.UUID
		paymentID uuid.UUID
		eventType entities.PaymentEventType
		eventData json.RawMessage
		createdAt time.Time
	)

	err := rows.Scan(
		&id,
		&paymentID,
		&eventType,
		&eventData,
		&createdAt,
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao fazer scan do evento de pagamento: %w", err)
	}

	return &entities.PaymentEvent{
		ID:        id,
		PaymentID: paymentID,
		EventType: eventType,
		EventData: eventData,
		CreatedAt: createdAt,
	}, nil
}
