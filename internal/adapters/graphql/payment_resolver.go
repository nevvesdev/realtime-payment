package graphql

import (
	"context"
	"strconv"

	"br.com.nevvesdev/realtime-payment/internal/adapters/graphql/generated"
	"br.com.nevvesdev/realtime-payment/internal/application/dto"
	"br.com.nevvesdev/realtime-payment/internal/infrastructure/logger"
	"github.com/google/uuid"
)

// Query

func (r *Resolver) Payment(ctx context.Context, id string) (*generated.Payment, error) {
	log := logger.Get()

	paymentID, err := uuid.Parse(id)
	if err != nil {
		log.WithField("error", err).Warn("ID de pagamento inválido")
		return nil, err
	}

	paymentOutput, err := r.paymentService.GetPayment(ctx, paymentID)
	if err != nil {
		log.WithField("error", err).Error("erro ao buscar pagamento")
		return nil, err
	}

	return paymentOutputToGraphQL(paymentOutput), nil
}

func (r *Resolver) Payments(ctx context.Context, accountID string, limit *int, offset *int) ([]*generated.Payment, error) {
	log := logger.Get()

	accountIDUUID, err := uuid.Parse(accountID)
	if err != nil {
		log.WithField("error", err).Warn("ID de conta inválido")
		return nil, err
	}

	limitVal := 20
	if limit != nil && *limit > 0 {
		limitVal = *limit
	}

	offsetVal := 0
	if offset != nil && *offset >= 0 {
		offsetVal = *offset
	}

	payments, err := r.paymentService.ListPayments(ctx, accountIDUUID, limitVal, offsetVal)
	if err != nil {
		log.WithField("error", err).Error("erro ao listar pagamentos")
		return nil, err
	}

	result := make([]*generated.Payment, len(payments))
	for i, p := range payments {
		result[i] = paymentOutputToGraphQL(p)
	}

	return result, nil
}

func (r *Resolver) PaymentEvents(ctx context.Context, paymentID string) ([]*generated.PaymentEvent, error) {
	log := logger.Get()

	paymentIDUUID, err := uuid.Parse(paymentID)
	if err != nil {
		log.WithField("error", err).Warn("ID de pagamento inválido")
		return nil, err
	}

	events, err := r.paymentService.GetPaymentEvents(ctx, paymentIDUUID, 100, 0)
	if err != nil {
		log.WithField("error", err).Error("erro ao buscar eventos de pagamento")
		return nil, err
	}

	result := make([]*generated.PaymentEvent, len(events))
	for i, e := range events {
		result[i] = paymentEventOutputToGraphQL(e)
	}

	return result, nil
}

// Mutation

func (r *Resolver) CreatePayment(ctx context.Context, input generated.CreatePaymentInput) (*generated.Payment, error) {
	log := logger.Get()

	accountID, err := uuid.Parse(input.AccountID)
	if err != nil {
		log.WithField("error", err).Warn("ID de conta inválido")
		return nil, err
	}

	amountFloat, err := strconv.ParseFloat(input.Amount, 64)
	if err != nil {
		log.WithField("error", err).Warn("valor de pagamento inválido")
		return nil, err
	}

	serviceInput := &dto.CreatePaymentInput{
		AccountID:      accountID,
		Amount:         amountFloat,
		Currency:       input.Currency,
		Description:    input.Description,
		IdempotencyKey: input.IdempotencyKey,
	}

	paymentOutput, err := r.paymentService.CreatePayment(ctx, serviceInput)
	if err != nil {
		log.WithField("error", err).Error("erro ao criar pagamento")
		return nil, err
	}

	// Publicar evento de subscription
	r.subscriptionMgr.PublishPaymentCreated(paymentOutput)

	return paymentOutputToGraphQL(paymentOutput), nil
}

func (r *Resolver) CancelPayment(ctx context.Context, id string) (*generated.Payment, error) {
	log := logger.Get()

	paymentID, err := uuid.Parse(id)
	if err != nil {
		log.WithField("error", err).Warn("ID de pagamento inválido")
		return nil, err
	}

	serviceInput := &dto.CancelPaymentInput{
		PaymentID: paymentID,
	}

	paymentOutput, err := r.paymentService.CancelPayment(ctx, serviceInput)
	if err != nil {
		log.WithField("error", err).Error("erro ao cancelar pagamento")
		return nil, err
	}

	// Publicar evento de subscription
	r.subscriptionMgr.PublishPaymentCancelled(paymentOutput)

	return paymentOutputToGraphQL(paymentOutput), nil
}

// Subscription

func (r *Resolver) PaymentStatusChanged(ctx context.Context, accountID string) (<-chan *generated.Payment, error) {
	log := logger.Get()

	accountIDUUID, err := uuid.Parse(accountID)
	if err != nil {
		log.WithField("error", err).Warn("ID de conta inválido")
		return nil, err
	}

	log.WithField("account_id", accountID).Info("subscription iniciada para mudanças de status de pagamento")

	return r.subscriptionMgr.SubscribePaymentStatusChanged(accountIDUUID), nil
}

func (r *Resolver) SettlementStatusChanged(ctx context.Context, paymentID string) (<-chan *generated.Settlement, error) {
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

func paymentOutputToGraphQL(output *dto.PaymentOutput) *generated.Payment {
	return &generated.Payment{
		ID:             output.ID,
		AccountID:      output.AccountID,
		Amount:         output.Amount,
		Currency:       output.Currency,
		Status:         generated.PaymentStatus(output.Status),
		Description:    &output.Description,
		IdempotencyKey: output.IdempotencyKey,
		CreatedAt:      output.CreatedAt,
		UpdatedAt:      output.UpdatedAt,
	}
}

func paymentEventOutputToGraphQL(output *dto.PaymentEventOutput) *generated.PaymentEvent {
	return &generated.PaymentEvent{
		ID:        output.ID,
		PaymentID: output.PaymentID,
		EventType: output.EventType,
		CreatedAt: output.CreatedAt,
	}
}
