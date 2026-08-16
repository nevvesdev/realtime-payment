package repositories

import (
	"context"

	"br.com.nevvesdev/realtime-payment/internal/domain/entities"
	"github.com/google/uuid"
)

type PaymentRepository interface {
	// Criar novo pagamento
	Create(ctx context.Context, payment *entities.Payment) error

	// Buscar pagamento por ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Payment, error)

	// Buscar pagamento por chave de idempotência
	FindByIdempotencyKey(ctx context.Context, key string) (*entities.Payment, error)

	// Listar pagamentos por conta
	FindByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*entities.Payment, error)

	// Atualizar status do pagamento
	Update(ctx context.Context, payment *entities.Payment) error

	// Contar pagamentos por conta
	CountByAccountID(ctx context.Context, accountID uuid.UUID) (int, error)

	// Deletar pagamento
	Delete(ctx context.Context, id uuid.UUID) error
}
