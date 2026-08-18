package graphql

import (
	"sync"

	"br.com.nevvesdev/realtime-payment/internal/adapters/graphql/generated"
	"br.com.nevvesdev/realtime-payment/internal/application/dto"
	"br.com.nevvesdev/realtime-payment/internal/infrastructure/logger"
	"github.com/google/uuid"
)

// SubscriptionManager gerencia subscriptions em tempo real via WebSocket
type SubscriptionManager struct {
	// Subscriptions de mudança de status de pagamento por conta
	paymentStatusChangedChannels map[uuid.UUID][]chan *generated.Payment
	paymentStatusMutex           sync.RWMutex

	// Subscriptions de mudança de status de liquidação por pagamento
	settlementStatusChangedChannels map[uuid.UUID][]chan *generated.Settlement
	settlementStatusMutex           sync.RWMutex
}

func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		paymentStatusChangedChannels:    make(map[uuid.UUID][]chan *generated.Payment),
		settlementStatusChangedChannels: make(map[uuid.UUID][]chan *generated.Settlement),
	}
}

// SubscribePaymentStatusChanged registra uma subscription para mudanças de status de pagamento
func (sm *SubscriptionManager) SubscribePaymentStatusChanged(accountID uuid.UUID) <-chan *generated.Payment {
	log := logger.Get()

	ch := make(chan *generated.Payment, 10) // Buffer de 10 eventos

	sm.paymentStatusMutex.Lock()
	sm.paymentStatusChangedChannels[accountID] = append(sm.paymentStatusChangedChannels[accountID], ch)
	sm.paymentStatusMutex.Unlock()

	log.WithField("account_id", accountID).Debug("subscriber adicionado para mudanças de pagamento")

	// Goroutine para cleanup quando o canal é fechado
	go func() {
		<-ch
		sm.unsubscribePaymentStatusChanged(accountID, ch)
	}()

	return ch
}

// SubscribeSettlementStatusChanged registra uma subscription para mudanças de status de liquidação
func (sm *SubscriptionManager) SubscribeSettlementStatusChanged(paymentID uuid.UUID) <-chan *generated.Settlement {
	log := logger.Get()

	ch := make(chan *generated.Settlement, 10) // Buffer de 10 eventos

	sm.settlementStatusMutex.Lock()
	sm.settlementStatusChangedChannels[paymentID] = append(sm.settlementStatusChangedChannels[paymentID], ch)
	sm.settlementStatusMutex.Unlock()

	log.WithField("payment_id", paymentID).Debug("subscriber adicionado para mudanças de liquidação")

	// Goroutine para cleanup quando o canal é fechado
	go func() {
		<-ch
		sm.unsubscribeSettlementStatusChanged(paymentID, ch)
	}()

	return ch
}

// PublishPaymentCreated publica um evento de pagamento criado
func (sm *SubscriptionManager) PublishPaymentCreated(payment *dto.PaymentOutput) {
	log := logger.Get()

	accountID, err := uuid.Parse(payment.AccountID)
	if err != nil {
		log.WithField("error", err).Error("erro ao fazer parse do ID de conta")
		return
	}

	sm.publishPaymentStatusChanged(accountID, payment)

	log.WithFields(map[string]interface{}{
		"account_id": accountID,
		"payment_id": payment.ID,
	}).Debug("evento de pagamento criado publicado")
}

// PublishPaymentCancelled publica um evento de pagamento cancelado
func (sm *SubscriptionManager) PublishPaymentCancelled(payment *dto.PaymentOutput) {
	log := logger.Get()

	accountID, err := uuid.Parse(payment.AccountID)
	if err != nil {
		log.WithField("error", err).Error("erro ao fazer parse do ID de conta")
		return
	}

	sm.publishPaymentStatusChanged(accountID, payment)

	log.WithFields(map[string]interface{}{
		"account_id": accountID,
		"payment_id": payment.ID,
	}).Debug("evento de pagamento cancelado publicado")
}

// PublishPaymentCompleted publica um evento de pagamento completado
func (sm *SubscriptionManager) PublishPaymentCompleted(payment *dto.PaymentOutput) {
	log := logger.Get()

	accountID, err := uuid.Parse(payment.AccountID)
	if err != nil {
		log.WithField("error", err).Error("erro ao fazer parse do ID de conta")
		return
	}

	sm.publishPaymentStatusChanged(accountID, payment)

	log.WithFields(map[string]interface{}{
		"account_id": accountID,
		"payment_id": payment.ID,
	}).Debug("evento de pagamento completado publicado")
}

// PublishSettlementCreated publica um evento de liquidação criada
func (sm *SubscriptionManager) PublishSettlementCreated(settlement *dto.SettlementOutput) {
	log := logger.Get()

	paymentID, err := uuid.Parse(settlement.PaymentID)
	if err != nil {
		log.WithField("error", err).Error("erro ao fazer parse do ID de pagamento")
		return
	}

	sm.publishSettlementStatusChanged(paymentID, settlement)

	log.WithFields(map[string]interface{}{
		"payment_id":    paymentID,
		"settlement_id": settlement.ID,
	}).Debug("evento de liquidação criada publicado")
}

// PublishSettlementCompleted publica um evento de liquidação concluída
func (sm *SubscriptionManager) PublishSettlementCompleted(settlement *dto.SettlementOutput) {
	log := logger.Get()

	paymentID, err := uuid.Parse(settlement.PaymentID)
	if err != nil {
		log.WithField("error", err).Error("erro ao fazer parse do ID de pagamento")
		return
	}

	sm.publishSettlementStatusChanged(paymentID, settlement)

	log.WithFields(map[string]interface{}{
		"payment_id":    paymentID,
		"settlement_id": settlement.ID,
	}).Debug("evento de liquidação concluída publicado")
}

// Private helpers

func (sm *SubscriptionManager) publishPaymentStatusChanged(accountID uuid.UUID, payment *dto.PaymentOutput) {
	sm.paymentStatusMutex.RLock()
	channels := sm.paymentStatusChangedChannels[accountID]
	sm.paymentStatusMutex.RUnlock()

	graphqlPayment := paymentOutputToGraphQL(payment)

	for _, ch := range channels {
		select {
		case ch <- graphqlPayment:
		default:
			// Canal cheio, pula este envio
		}
	}
}

func (sm *SubscriptionManager) publishSettlementStatusChanged(paymentID uuid.UUID, settlement *dto.SettlementOutput) {
	sm.settlementStatusMutex.RLock()
	channels := sm.settlementStatusChangedChannels[paymentID]
	sm.settlementStatusMutex.RUnlock()

	graphqlSettlement := settlementOutputToGraphQL(settlement)

	for _, ch := range channels {
		select {
		case ch <- graphqlSettlement:
		default:
			// Canal cheio, pula este envio
		}
	}
}

func (sm *SubscriptionManager) unsubscribePaymentStatusChanged(accountID uuid.UUID, ch chan *generated.Payment) {
	log := logger.Get()

	sm.paymentStatusMutex.Lock()
	defer sm.paymentStatusMutex.Unlock()

	channels := sm.paymentStatusChangedChannels[accountID]
	for i, c := range channels {
		if c == ch {
			sm.paymentStatusChangedChannels[accountID] = append(channels[:i], channels[i+1:]...)
			close(ch)
			break
		}
	}

	if len(sm.paymentStatusChangedChannels[accountID]) == 0 {
		delete(sm.paymentStatusChangedChannels, accountID)
	}

	log.WithField("account_id", accountID).Debug("subscriber removido para mudanças de pagamento")
}

func (sm *SubscriptionManager) unsubscribeSettlementStatusChanged(paymentID uuid.UUID, ch chan *generated.Settlement) {
	log := logger.Get()

	sm.settlementStatusMutex.Lock()
	defer sm.settlementStatusMutex.Unlock()

	channels := sm.settlementStatusChangedChannels[paymentID]
	for i, c := range channels {
		if c == ch {
			sm.settlementStatusChangedChannels[paymentID] = append(channels[:i], channels[i+1:]...)
			close(ch)
			break
		}
	}

	if len(sm.settlementStatusChangedChannels[paymentID]) == 0 {
		delete(sm.settlementStatusChangedChannels, paymentID)
	}

	log.WithField("payment_id", paymentID).Debug("subscriber removido para mudanças de liquidação")
}
