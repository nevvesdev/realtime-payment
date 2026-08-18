package graphql

import (
	"context"
	"time"

	"br.com.nevvesdev/realtime-payment/internal/adapters/graphql/generated"
	"br.com.nevvesdev/realtime-payment/internal/application/dto"
	"br.com.nevvesdev/realtime-payment/internal/infrastructure/logger"
	"github.com/google/uuid"
)

// Query

func (r *Resolver) Settlement(ctx context.Context, id string) (*generated.Settlement, error) {
	log := logger.Get()

	settlementID, err := uuid.Parse(id)
	if err != nil {
		log.WithField("error", err).Warn("ID de liquidação inválido")
		return nil, err
	}

	settlementOutput, err := r.settlementService.GetSettlement(ctx, settlementID)
	if err != nil {
		log.WithField("error", err).Error("erro ao buscar liquidação")
		return nil, err
	}

	return settlementOutputToGraphQL(settlementOutput), nil
}

// Mutation

func (r *Resolver) CreateSettlement(ctx context.Context, paymentID string, settlementDate time.Time) (*generated.Settlement, error) {
	log := logger.Get()

	paymentIDUUID, err := uuid.Parse(paymentID)
	if err != nil {
		log.WithField("error", err).Warn("ID de pagamento inválido")
		return nil, err
	}

	serviceInput := &dto.CreateSettlementInput{
		PaymentID:      paymentIDUUID,
		SettlementDate: settlementDate,
	}

	settlementOutput, err := r.settlementService.CreateSettlement(ctx, serviceInput)
	if err != nil {
		log.WithField("error", err).Error("erro ao criar liquidação")
		return nil, err
	}

	// Publicar evento de subscription
	r.subscriptionMgr.PublishSettlementCreated(settlementOutput)

	return settlementOutputToGraphQL(settlementOutput), nil
}

// Subscription

func (r *Resolver) SettlementStatusChangedResolver(ctx context.Context, paymentID string) (<-chan *generated.Settlement, error) {
	log := logger.Get()

	paymentIDUUID, err := uuid.Parse(paymentID)
	if err != nil {
		log.WithField("error", err).Warn("ID de pagamento inválido")
		return nil, err
	}

	log.WithField("payment_id", paymentID).Info("subscription iniciada para mudanças de status de liquidação")

	return r.subscriptionMgr.SubscribeSettlementStatusChanged(paymentIDUUID), nil
}

// Helpers

func settlementOutputToGraphQL(output *dto.SettlementOutput) *generated.Settlement {
	return &generated.Settlement{
		ID:             output.ID,
		PaymentID:      output.PaymentID,
		SettlementDate: output.SettlementDate,
		Status:         generated.SettlementStatus(output.Status),
		CreatedAt:      output.CreatedAt,
		UpdatedAt:      output.UpdatedAt,
	}
}
