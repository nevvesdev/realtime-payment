package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"br.com.nevvesdev/realtime-payment/internal/domain/entities"
	domainerrors "br.com.nevvesdev/realtime-payment/internal/domain/errors"
	"br.com.nevvesdev/realtime-payment/internal/domain/repositories"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPaymentRepository(pool *pgxpool.Pool) repositories.PaymentRepository {
	return &PostgresPaymentRepository{pool: pool}
}

func (r *PostgresPaymentRepository) Create(ctx context.Context, payment *entities.Payment) error {
	query := `
		INSERT INTO payments (id, account_id, amount, currency, status, description, idempotency_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		payment.ID,
		payment.AccountID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.Description,
		payment.IdempotencyKey,
		payment.CreatedAt,
		payment.UpdatedAt,
	)

	if err != nil {
		return domainerrors.NewDomainErrorWithCause(
			domainerrors.ErrDatabaseError,
			"erro ao criar pagamento",
			err,
		)
	}

	return nil
}

func (r *PostgresPaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Payment, error) {
	query := `
		SELECT id, account_id, amount, currency, status, description, idempotency_key, created_at, updated_at
		FROM payments
		WHERE id = $1
	`

	payment, err := r.scanPayment(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.NewDomainError(domainerrors.ErrPaymentNotFound, "pagamento não encontrado")
		}
		return nil, domainerrors.NewDomainErrorWithCause(
			domainerrors.ErrDatabaseError,
			"erro ao buscar pagamento",
			err,
		)
	}

	return payment, nil
}

func (r *PostgresPaymentRepository) FindByIdempotencyKey(ctx context.Context, key string) (*entities.Payment, error) {
	query := `
		SELECT id, account_id, amount, currency, status, description, idempotency_key, created_at, updated_at
		FROM payments
		WHERE idempotency_key = $1
	`

	payment, err := r.scanPayment(r.pool.QueryRow(ctx, query, key))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.NewDomainError(domainerrors.ErrPaymentNotFound, "pagamento não encontrado")
		}
		return nil, domainerrors.NewDomainErrorWithCause(
			domainerrors.ErrDatabaseError,
			"erro ao buscar pagamento por chave de idempotência",
			err,
		)
	}

	return payment, nil
}

func (r *PostgresPaymentRepository) FindByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*entities.Payment, error) {
	query := `
		SELECT id, account_id, amount, currency, status, description, idempotency_key, created_at, updated_at
		FROM payments
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, accountID, limit, offset)
	if err != nil {
		return nil, domainerrors.NewDomainErrorWithCause(
			domainerrors.ErrDatabaseError,
			"erro ao buscar pagamentos da conta",
			err,
		)
	}
	defer rows.Close()

	var payments []*entities.Payment
	for rows.Next() {
		payment, err := r.scanPaymentFromRows(rows)
		if err != nil {
			return nil, domainerrors.NewDomainErrorWithCause(
				domainerrors.ErrDatabaseError,
				"erro ao processar linha de pagamento",
				err,
			)
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

func (r *PostgresPaymentRepository) Update(ctx context.Context, payment *entities.Payment) error {
	query := `
		UPDATE payments
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.pool.Exec(ctx, query, payment.Status, payment.UpdatedAt, payment.ID)
	if err != nil {
		return domainerrors.NewDomainErrorWithCause(
			domainerrors.ErrDatabaseError,
			"erro ao atualizar pagamento",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return domainerrors.NewDomainError(domainerrors.ErrPaymentNotFound, "pagamento não encontrado para atualização")
	}

	return nil
}

func (r *PostgresPaymentRepository) CountByAccountID(ctx context.Context, accountID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM payments WHERE account_id = $1`

	err := r.pool.QueryRow(ctx, query, accountID).Scan(&count)
	if err != nil {
		return 0, domainerrors.NewDomainErrorWithCause(
			domainerrors.ErrDatabaseError,
			"erro ao contar pagamentos",
			err,
		)
	}

	return count, nil
}

func (r *PostgresPaymentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM payments WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return domainerrors.NewDomainErrorWithCause(
			domainerrors.ErrDatabaseError,
			"erro ao deletar pagamento",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return domainerrors.NewDomainError(domainerrors.ErrPaymentNotFound, "pagamento não encontrado para deleção")
	}

	return nil
}

func (r *PostgresPaymentRepository) scanPayment(row interface {
	Scan(dest ...interface{}) error
}) (*entities.Payment, error) {
	var (
		id             uuid.UUID
		accountID      uuid.UUID
		amount         float64
		currency       string
		status         entities.PaymentStatus
		description    sql.NullString
		idempotencyKey string
		createdAt      time.Time
		updatedAt      time.Time
	)

	err := row.Scan(
		&id,
		&accountID,
		&amount,
		&currency,
		&status,
		&description,
		&idempotencyKey,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao fazer scan do pagamento: %w", err)
	}

	return &entities.Payment{
		ID:             id,
		AccountID:      accountID,
		Amount:         amount,
		Currency:       currency,
		Status:         status,
		Description:    description.String,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}, nil
}

func (r *PostgresPaymentRepository) scanPaymentFromRows(rows interface {
	Scan(dest ...interface{}) error
}) (*entities.Payment, error) {
	return r.scanPayment(rows)
}
