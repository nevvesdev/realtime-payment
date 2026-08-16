package repositories

import (
	"context"
	"time"

	"br.com.nevvesdev/realtime-payment/internal/domain/entities"
	"github.com/google/uuid"
)

type SettlementRepository interface {
	// Criar nova liquidação
	Create(ctx context.Context, settlement *entities.Settlement) error

	// Buscar liquidação por ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Settlement, error)

	// Buscar liquidação por ID de pagamento
	FindByPaymentID(ctx context.Context, paymentID uuid.UUID) (*entities.Settlement, error)

	// Listar liquidações por data
	FindBySettlementDate(ctx context.Context, date time.Time, limit, offset int) ([]*entities.Settlement, error)

	// Listar liquidações pendentes
	FindPending(ctx context.Context, limit, offset int) ([]*entities.Settlement, error)

	// Atualizar liquidação
	Update(ctx context.Context, settlement *entities.Settlement) error

	// Deletar liquidação
	Delete(ctx context.Context, id uuid.UUID) error
}
