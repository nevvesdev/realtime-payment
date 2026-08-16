package services

import (
	"context"
	"fmt"
	"time"

	"br.com.nevvesdev/realtime-payment/internal/application/dto"
	"br.com.nevvesdev/realtime-payment/internal/domain/entities"
	domainerrors "br.com.nevvesdev/realtime-payment/internal/domain/errors"
	"br.com.nevvesdev/realtime-payment/internal/domain/repositories"
	"br.com.nevvesdev/realtime-payment/internal/infrastructure/logger"
	"github.com/google/uuid"
)

type SettlementService struct {
	settlementRepo   repositories.SettlementRepository
	paymentRepo      repositories.PaymentRepository
	paymentEventRepo repositories.PaymentEventRepository
	outboxRepo       repositories.OutboxRepository
}

func NewSettlementService(
	settlementRepo repositories.SettlementRepository,
	paymentRepo repositories.PaymentRepository,
	paymentEventRepo repositories.PaymentEventRepository,
	outboxRepo repositories.OutboxRepository,
) *SettlementService {
	return &SettlementService{
		settlementRepo:   settlementRepo,
		paymentRepo:      paymentRepo,
		paymentEventRepo: paymentEventRepo,
		outboxRepo:       outboxRepo,
	}
}

// CreateSettlement orquestra a criação de uma liquidação
func (s *SettlementService) CreateSettlement(ctx context.Context, input *dto.CreateSettlementInput) (*dto.SettlementOutput, error) {
	log := logger.Get()

	// Validar entrada
	if err := validateCreateSettlementInput(input); err != nil {
		log.WithField("error", err).Warn("validação falhou para criar liquidação")
		return nil, err
	}

	// Verificar se pagamento existe e está completo
	payment, err := s.paymentRepo.FindByID(ctx, input.PaymentID)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"payment_id": input.PaymentID,
			"error":      err,
		}).Error("erro ao buscar pagamento para liquidação")
		return nil, err
	}

	if payment.Status != entities.PaymentStatusCompleted {
		log.WithFields(map[string]interface{}{
			"payment_id": input.PaymentID,
			"status":     payment.Status,
		}).Warn("pagamento deve estar completo para liquidação")
		return nil, fmt.Errorf("pagamento deve estar completo para liquidação, status atual: %s", payment.Status)
	}

	// Criar entidade de liquidação
	settlement := entities.NewSettlement(input.PaymentID, input.SettlementDate)

	// Salvar no repositório
	if err := s.settlementRepo.Create(ctx, settlement); err != nil {
		log.WithFields(map[string]interface{}{
			"payment_id": input.PaymentID,
			"error":      err,
		}).Error("erro ao criar liquidação no repositório")
		return nil, err
	}

	// Criar evento de domínio
	event, err := entities.NewPaymentEvent(
		payment.ID,
		entities.EventTypeSettlementCreated,
		map[string]interface{}{
			"settlement_id": settlement.ID,
			"payment_id":    payment.ID,
			"amount":        payment.Amount,
		},
	)
	if err != nil {
		log.WithField("error", err).Error("erro ao criar evento de liquidação criada")
		return nil, err
	}

	// Salvar evento
	if err := s.paymentEventRepo.Save(ctx, event); err != nil {
		log.WithField("error", err).Error("erro ao salvar evento de liquidação criada")
		return nil, err
	}

	// Salvar na outbox
	if err := s.saveEventToOutbox(ctx, payment.ID, event); err != nil {
		log.WithField("error", err).Error("erro ao salvar evento de liquidação na outbox")
		return nil, err
	}

	log.WithFields(map[string]interface{}{
		"settlement_id": settlement.ID,
		"payment_id":    payment.ID,
	}).Info("liquidação criada com sucesso")

	return dto.NewSettlementOutput(settlement), nil
}

// GetSettlement recupera uma liquidação pelo ID
func (s *SettlementService) GetSettlement(ctx context.Context, settlementID uuid.UUID) (*dto.SettlementOutput, error) {
	log := logger.Get()

	settlement, err := s.settlementRepo.FindByID(ctx, settlementID)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"settlement_id": settlementID,
			"error":         err,
		}).Error("erro ao buscar liquidação")
		return nil, err
	}

	return dto.NewSettlementOutput(settlement), nil
}

// GetSettlementByPaymentID recupera liquidação pelo ID de pagamento
func (s *SettlementService) GetSettlementByPaymentID(ctx context.Context, paymentID uuid.UUID) (*dto.SettlementOutput, error) {
	log := logger.Get()

	settlement, err := s.settlementRepo.FindByPaymentID(ctx, paymentID)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"payment_id": paymentID,
			"error":      err,
		}).Error("erro ao buscar liquidação por pagamento")
		return nil, err
	}

	return dto.NewSettlementOutput(settlement), nil
}

// ListSettlementsByDate lista liquidações por data
func (s *SettlementService) ListSettlementsByDate(ctx context.Context, date time.Time, limit, offset int) ([]*dto.SettlementOutput, error) {
	log := logger.Get()

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	settlements, err := s.settlementRepo.FindBySettlementDate(ctx, date, limit, offset)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"date":  date,
			"error": err,
		}).Error("erro ao listar liquidações por data")
		return nil, err
	}

	return dto.NewSettlementOutputList(settlements), nil
}

// ListPendingSettlements lista liquidações pendentes
func (s *SettlementService) ListPendingSettlements(ctx context.Context, limit, offset int) ([]*dto.SettlementOutput, error) {
	log := logger.Get()

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	settlements, err := s.settlementRepo.FindPending(ctx, limit, offset)
	if err != nil {
		log.WithField("error", err).Error("erro ao listar liquidações pendentes")
		return nil, err
	}

	return dto.NewSettlementOutputList(settlements), nil
}

// CompleteSettlement marca uma liquidação como concluída
func (s *SettlementService) CompleteSettlement(ctx context.Context, settlementID uuid.UUID) (*dto.SettlementOutput, error) {
	log := logger.Get()

	settlement, err := s.settlementRepo.FindByID(ctx, settlementID)
	if err != nil {
		log.WithField("error", err).Error("erro ao buscar liquidação para completar")
		return nil, err
	}

	if !settlement.CanTransitionTo(entities.SettlementStatusSettled) {
		log.Warn("não é possível completar liquidação neste estado")
		return nil, fmt.Errorf("não é possível completar liquidação com status %s", settlement.Status)
	}

	if err := settlement.TransitionTo(entities.SettlementStatusSettled); err != nil {
		return nil, err
	}

	if err := s.settlementRepo.Update(ctx, settlement); err != nil {
		log.WithField("error", err).Error("erro ao atualizar liquidação concluída")
		return nil, err
	}

	event, err := entities.NewPaymentEvent(
		settlement.PaymentID,
		entities.EventTypeSettlementCompleted,
		map[string]interface{}{
			"settlement_id": settlement.ID,
			"payment_id":    settlement.PaymentID,
		},
	)
	if err != nil {
		log.WithField("error", err).Error("erro ao criar evento de liquidação concluída")
		return nil, err
	}

	if err := s.paymentEventRepo.Save(ctx, event); err != nil {
		log.WithField("error", err).Error("erro ao salvar evento de liquidação concluída")
		return nil, err
	}

	if err := s.saveEventToOutbox(ctx, settlement.PaymentID, event); err != nil {
		log.WithField("error", err).Error("erro ao salvar evento de liquidação na outbox")
		return nil, err
	}

	log.WithField("settlement_id", settlement.ID).Info("liquidação concluída com sucesso")

	return dto.NewSettlementOutput(settlement), nil
}

// Helper para salvar evento na outbox
func (s *SettlementService) saveEventToOutbox(ctx context.Context, paymentID uuid.UUID, event *entities.PaymentEvent) error {
	outboxEvent := &repositories.OutboxEvent{
		AggregateID: paymentID,
		EventType:   string(event.EventType),
		EventData:   event.EventData,
		Processed:   false,
	}

	return s.outboxRepo.Save(ctx, outboxEvent)
}

// validateCreateSettlementInput valida dados de entrada
func validateCreateSettlementInput(input *dto.CreateSettlementInput) error {
	if input.SettlementDate.IsZero() {
		return domainerrors.NewDomainError(
			domainerrors.ErrInvalidInput,
			"data de liquidação é obrigatória",
		)
	}

	if input.SettlementDate.Before(time.Now()) {
		return domainerrors.NewDomainError(
			domainerrors.ErrInvalidInput,
			"data de liquidação não pode estar no passado",
		)
	}

	return nil
}
