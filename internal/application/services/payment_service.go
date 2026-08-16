package services

import (
	"context"
	"fmt"

	"br.com.nevvesdev/realtime-payment/internal/application/dto"
	"br.com.nevvesdev/realtime-payment/internal/domain/entities"
	domainerrors "br.com.nevvesdev/realtime-payment/internal/domain/errors"
	"br.com.nevvesdev/realtime-payment/internal/domain/repositories"
	"br.com.nevvesdev/realtime-payment/internal/infrastructure/logger"
	"github.com/google/uuid"
)

type PaymentService struct {
	paymentRepo      repositories.PaymentRepository
	settlementRepo   repositories.SettlementRepository
	paymentEventRepo repositories.PaymentEventRepository
	outboxRepo       repositories.OutboxRepository
}

func NewPaymentService(
	paymentRepo repositories.PaymentRepository,
	settlementRepo repositories.SettlementRepository,
	paymentEventRepo repositories.PaymentEventRepository,
	outboxRepo repositories.OutboxRepository,
) *PaymentService {
	return &PaymentService{
		paymentRepo:      paymentRepo,
		settlementRepo:   settlementRepo,
		paymentEventRepo: paymentEventRepo,
		outboxRepo:       outboxRepo,
	}
}

// CreatePayment orquestra a criação de um novo pagamento
func (s *PaymentService) CreatePayment(ctx context.Context, input *dto.CreatePaymentInput) (*dto.PaymentOutput, error) {
	log := logger.Get()

	// Validar entrada
	if err := validateCreatePaymentInput(input); err != nil {
		log.WithField("error", err).Warn("validação falhou para criar pagamento")
		return nil, err
	}

	// Verificar se pagamento com mesma chave de idempotência já existe
	existingPayment, err := s.paymentRepo.FindByIdempotencyKey(ctx, input.IdempotencyKey)
	if err == nil && existingPayment != nil {
		log.WithFields(map[string]interface{}{
			"idempotency_key": input.IdempotencyKey,
			"payment_id":      existingPayment.ID,
		}).Info("pagamento já existe com essa chave de idempotência")
		return dto.NewPaymentOutput(existingPayment), nil
	}

	if err != nil && !domainerrors.IsErrorCode(err, domainerrors.ErrPaymentNotFound) {
		log.WithField("error", err).Error("erro ao verificar chave de idempotência")
		return nil, err
	}

	// Criar entidade de pagamento
	payment := entities.NewPayment(
		input.AccountID,
		input.Amount,
		input.Currency,
		input.Description,
		input.IdempotencyKey,
	)

	// Salvar no repositório (transaction)
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		log.WithFields(map[string]interface{}{
			"account_id": input.AccountID,
			"amount":     input.Amount,
			"error":      err,
		}).Error("erro ao criar pagamento no repositório")
		return nil, err
	}

	// Criar evento de domínio
	event, err := entities.NewPaymentEvent(
		payment.ID,
		entities.EventTypePaymentCreated,
		map[string]interface{}{
			"payment_id": payment.ID,
			"amount":     payment.Amount,
			"currency":   payment.Currency,
			"account_id": payment.AccountID,
		},
	)
	if err != nil {
		log.WithField("error", err).Error("erro ao criar evento de pagamento criado")
		return nil, err
	}

	// Salvar evento (event sourcing)
	if err := s.paymentEventRepo.Save(ctx, event); err != nil {
		log.WithField("error", err).Error("erro ao salvar evento de pagamento criado")
		return nil, err
	}

	// Salvar na outbox para publicação assíncrona
	if err := s.saveEventToOutbox(ctx, payment.ID, event); err != nil {
		log.WithField("error", err).Error("erro ao salvar evento na outbox")
		return nil, err
	}

	log.WithFields(map[string]interface{}{
		"payment_id": payment.ID,
		"account_id": payment.AccountID,
	}).Info("pagamento criado com sucesso")

	return dto.NewPaymentOutput(payment), nil
}

// GetPayment recupera um pagamento pelo ID
func (s *PaymentService) GetPayment(ctx context.Context, paymentID uuid.UUID) (*dto.PaymentOutput, error) {
	log := logger.Get()

	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"payment_id": paymentID,
			"error":      err,
		}).Error("erro ao buscar pagamento")
		return nil, err
	}

	return dto.NewPaymentOutput(payment), nil
}

// ListPayments lista pagamentos de uma conta com paginação
func (s *PaymentService) ListPayments(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*dto.PaymentOutput, error) {
	log := logger.Get()

	// Validar paginação
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	payments, err := s.paymentRepo.FindByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"account_id": accountID,
			"error":      err,
		}).Error("erro ao listar pagamentos")
		return nil, err
	}

	return dto.NewPaymentOutputList(payments), nil
}

// CancelPayment cancela um pagamento existente
func (s *PaymentService) CancelPayment(ctx context.Context, input *dto.CancelPaymentInput) (*dto.PaymentOutput, error) {
	log := logger.Get()

	// Buscar pagamento
	payment, err := s.paymentRepo.FindByID(ctx, input.PaymentID)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"payment_id": input.PaymentID,
			"error":      err,
		}).Error("erro ao buscar pagamento para cancelamento")
		return nil, err
	}

	// Validar transição de estado
	if !payment.CanTransitionTo(entities.PaymentStatusCancelled) {
		log.WithFields(map[string]interface{}{
			"payment_id":     input.PaymentID,
			"current_status": payment.Status,
		}).Warn("não é possível cancelar pagamento neste estado")
		return nil, fmt.Errorf("não é possível cancelar pagamento com status %s", payment.Status)
	}

	// Transicionar estado
	if err := payment.TransitionTo(entities.PaymentStatusCancelled); err != nil {
		log.WithField("error", err).Error("erro ao transicionar estado do pagamento")
		return nil, err
	}

	// Atualizar no repositório
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		log.WithField("error", err).Error("erro ao atualizar pagamento cancelado")
		return nil, err
	}

	// Criar evento de domínio
	event, err := entities.NewPaymentEvent(
		payment.ID,
		entities.EventTypePaymentCancelled,
		map[string]interface{}{
			"payment_id": payment.ID,
			"reason":     "cancelado pelo usuário",
		},
	)
	if err != nil {
		log.WithField("error", err).Error("erro ao criar evento de pagamento cancelado")
		return nil, err
	}

	// Salvar evento
	if err := s.paymentEventRepo.Save(ctx, event); err != nil {
		log.WithField("error", err).Error("erro ao salvar evento de pagamento cancelado")
		return nil, err
	}

	// Salvar na outbox
	if err := s.saveEventToOutbox(ctx, payment.ID, event); err != nil {
		log.WithField("error", err).Error("erro ao salvar evento cancelado na outbox")
		return nil, err
	}

	log.WithFields(map[string]interface{}{
		"payment_id": payment.ID,
	}).Info("pagamento cancelado com sucesso")

	return dto.NewPaymentOutput(payment), nil
}

// ProcessPayment marca um pagamento como processando
func (s *PaymentService) ProcessPayment(ctx context.Context, paymentID uuid.UUID) (*dto.PaymentOutput, error) {
	log := logger.Get()

	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		log.WithField("error", err).Error("erro ao buscar pagamento para processar")
		return nil, err
	}

	if !payment.CanTransitionTo(entities.PaymentStatusProcessing) {
		log.Warn("não é possível processar pagamento neste estado")
		return nil, fmt.Errorf("não é possível processar pagamento com status %s", payment.Status)
	}

	if err := payment.TransitionTo(entities.PaymentStatusProcessing); err != nil {
		return nil, err
	}

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, err
	}

	event, err := entities.NewPaymentEvent(
		payment.ID,
		entities.EventTypePaymentProcessing,
		map[string]interface{}{"payment_id": payment.ID},
	)
	if err != nil {
		return nil, err
	}

	if err := s.paymentEventRepo.Save(ctx, event); err != nil {
		return nil, err
	}

	if err := s.saveEventToOutbox(ctx, payment.ID, event); err != nil {
		return nil, err
	}

	log.WithField("payment_id", payment.ID).Info("pagamento marcado como processando")

	return dto.NewPaymentOutput(payment), nil
}

// CompletePayment marca um pagamento como concluído
func (s *PaymentService) CompletePayment(ctx context.Context, paymentID uuid.UUID) (*dto.PaymentOutput, error) {
	log := logger.Get()

	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		log.WithField("error", err).Error("erro ao buscar pagamento para completar")
		return nil, err
	}

	if !payment.CanTransitionTo(entities.PaymentStatusCompleted) {
		log.Warn("não é possível completar pagamento neste estado")
		return nil, fmt.Errorf("não é possível completar pagamento com status %s", payment.Status)
	}

	if err := payment.TransitionTo(entities.PaymentStatusCompleted); err != nil {
		return nil, err
	}

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, err
	}

	event, err := entities.NewPaymentEvent(
		payment.ID,
		entities.EventTypePaymentCompleted,
		map[string]interface{}{"payment_id": payment.ID},
	)
	if err != nil {
		return nil, err
	}

	if err := s.paymentEventRepo.Save(ctx, event); err != nil {
		return nil, err
	}

	if err := s.saveEventToOutbox(ctx, payment.ID, event); err != nil {
		return nil, err
	}

	log.WithField("payment_id", payment.ID).Info("pagamento completado com sucesso")

	return dto.NewPaymentOutput(payment), nil
}

// Helper para salvar evento na outbox
func (s *PaymentService) saveEventToOutbox(ctx context.Context, paymentID uuid.UUID, event *entities.PaymentEvent) error {
	outboxEvent := &repositories.OutboxEvent{
		AggregateID: paymentID,
		EventType:   string(event.EventType),
		EventData:   event.EventData,
		Processed:   false,
	}

	return s.outboxRepo.Save(ctx, outboxEvent)
}

// GetPaymentEvents retorna histórico de eventos de um pagamento
func (s *PaymentService) GetPaymentEvents(ctx context.Context, paymentID uuid.UUID, limit, offset int) ([]*dto.PaymentEventOutput, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	events, err := s.paymentEventRepo.FindByPaymentIDWithPagination(ctx, paymentID, limit, offset)
	if err != nil {
		return nil, err
	}

	return dto.NewPaymentEventOutputList(events), nil
}

// validateCreatePaymentInput valida os dados de entrada
func validateCreatePaymentInput(input *dto.CreatePaymentInput) error {
	if input.Amount <= 0 {
		return domainerrors.NewDomainError(
			domainerrors.ErrInvalidInput,
			"o valor do pagamento deve ser positivo",
		)
	}

	if input.Currency == "" {
		return domainerrors.NewDomainError(
			domainerrors.ErrInvalidInput,
			"moeda é obrigatória",
		)
	}

	if input.IdempotencyKey == "" {
		return domainerrors.NewDomainError(
			domainerrors.ErrInvalidInput,
			"chave de idempotência é obrigatória",
		)
	}

	return nil
}
