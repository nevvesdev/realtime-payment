package repositories

import (
	"context"

	"br.com.nevvesdev/realtime-payment/internal/domain/entities"
	"github.com/google/uuid"
)

type PaymentEventRepository interface {
	// Salvar evento de pagamento
	Save(ctx context.Context, event *entities.PaymentEvent) error

	// Buscar eventos por ID de pagamento
	FindByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*entities.PaymentEvent, error)

	// Buscar eventos com paginação
	FindByPaymentIDWithPagination(
		ctx context.Context,
		paymentID uuid.UUID,
		limit, offset int,
	) ([]*entities.PaymentEvent, error)

	// Contar eventos por pagamento
	CountByPaymentID(ctx context.Context, paymentID uuid.UUID) (int, error)
}
