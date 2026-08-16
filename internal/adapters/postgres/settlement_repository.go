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

type PostgresSettlementRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresSettlementRepository(pool *pgxpool.Pool) repositories.SettlementRepository {
	return &PostgresSettlementRepository{pool: pool}
}

func (r *PostgresSettlementRepository) Create(ctx context.Context, settlement *entities.Settlement) error {
	query := `
		INSERT INTO settlements (id, payment_id, settlement_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		settlement.ID,
		settlement.PaymentID,
		settlement.SettlementDate,
		settlement.Status,
		settlement.CreatedAt,
		settlement.UpdatedAt,
	)

	if err != nil {
		return domainerrors.NewDomainErrorWithCause(
			domainerrors.ErrDatabaseError,
			"erro ao criar liquidação",
			err,
		)
	}

	return nil
}

func (r *PostgresSettlementRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Settlement, error) {
	query := `
		SELECT id, payment_id, settlement_date, status, created_at, updated_at
		FROM settlements
		WHERE id = $1
	`

	settlement, err := r.scanSettlement(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.NewDomainError(domainerrors.ErrSettlementNotFound, "liquidação não encontrada")
		}
		return nil, domainerrors.NewDomainErrorWithCause(
			domainerrors.ErrDatabaseError,
			"erro ao buscar liquidação",
			err,
		)
	}

	return settlement, nil
}

func (r *PostgresSettlementRepository) FindByPaymentID(ctx context.Context, paymentID uuid.UUID) (*entities.Settlement, error) {
	query := `
		SELECT id, payment_id, settlement_date, status, created_at, updated_at
		FROM settlements
		WHERE payment_id = $1
	`

	settlement, err := r.scanSettlement(r.pool.QueryRow(ctx, query, paymentID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.NewDomainError(domainerrors.ErrSettlementNotFound, "liquidação não encontrada")
		}
		return nil, domainerrors.NewDomainErrorWithCause(
			domainerrors.ErrDatabaseError,
			"erro ao buscar liquidação por pagamento",
			err,
		)
	}

	return settlement, nil
}

func (r *PostgresSettlementRepository) FindBySettlementDate(ctx context.Context, date time.Time, limit, offset int) ([]*entities.Settlement, error) {
	query := `
		SELECT id, payment_id, settlement_date, status, created_at, updated_at
		FROM settlements
		WHERE DATE(settlement_date) = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, date, limit, offset)
	if err != nil {
		return nil, domainerrors.NewDomainErrorWithCause(
			domainerrors.ErrDatabaseError,
			"erro ao buscar liquidações por data",
			err,
		)
	}
	defer rows.Close()

	var settlements []*entities.Settlement
	for rows.Next() {
		settlement, err := r.scanSettlementFromRows(rows)
		if err != nil {
			return nil, domainerrors.NewDomainErrorWithCause(
				domainerrors.ErrDatabaseError,
				"erro ao processar linha de liquidação",
				err,
			)
		}
		settlements = append(settlements, settlement)
	}

	return settlements, nil
}

func (r *PostgresSettlementRepository) FindPending(ctx context.Context, limit, offset int) ([]*entities.Settlement, error) {
	query := `
		SELECT id, payment_id, settlement_date, status, created_at, updated_at
		FROM settlements
		WHERE status = $1
		ORDER BY settlement_date ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, entities.SettlementStatusPending, limit, offset)
	if err != nil {
		return nil, domainerrors.NewDomainErrorWithCause(
			domainerrors.ErrDatabaseError,
			"erro ao buscar liquidações pendentes",
			err,
		)
	}
	defer rows.Close()

	var settlements []*entities.Settlement
	for rows.Next() {
		settlement, err := r.scanSettlementFromRows(rows)
		if err != nil {
			return nil, domainerrors.NewDomainErrorWithCause(
				domainerrors.ErrDatabaseError,
				"erro ao processar linha de liquidação pendente",
				err,
			)
		}
		settlements = append(settlements, settlement)
	}

	return settlements, nil
}

func (r *PostgresSettlementRepository) Update(ctx context.Context, settlement *entities.Settlement) error {
	query := `
		UPDATE settlements
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.pool.Exec(ctx, query, settlement.Status, settlement.UpdatedAt, settlement.ID)
	if err != nil {
		return domainerrors.NewDomainErrorWithCause(
			domainerrors.ErrDatabaseError,
			"erro ao atualizar liquidação",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return domainerrors.NewDomainError(domainerrors.ErrSettlementNotFound, "liquidação não encontrada para atualização")
	}

	return nil
}

func (r *PostgresSettlementRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM settlements WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return domainerrors.NewDomainErrorWithCause(
			domainerrors.ErrDatabaseError,
			"erro ao deletar liquidação",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return domainerrors.NewDomainError(domainerrors.ErrSettlementNotFound, "liquidação não encontrada para deleção")
	}

	return nil
}

func (r *PostgresSettlementRepository) scanSettlement(row interface {
	Scan(dest ...interface{}) error
}) (*entities.Settlement, error) {
	var (
		id             uuid.UUID
		paymentID      uuid.UUID
		settlementDate time.Time
		status         entities.SettlementStatus
		createdAt      time.Time
		updatedAt      time.Time
	)

	err := row.Scan(
		&id,
		&paymentID,
		&settlementDate,
		&status,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao fazer scan da liquidação: %w", err)
	}

	return &entities.Settlement{
		ID:             id,
		PaymentID:      paymentID,
		SettlementDate: settlementDate,
		Status:         status,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}, nil
}

func (r *PostgresSettlementRepository) scanSettlementFromRows(rows interface {
	Scan(dest ...interface{}) error
}) (*entities.Settlement, error) {
	return r.scanSettlement(rows)
}
