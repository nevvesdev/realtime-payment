package dto

import (
	"time"

	"br.com.nevvesdev/realtime-payment/internal/domain/entities"
	"github.com/google/uuid"
)

// CreatePaymentInput recebe dados de entrada para criar um pagamento
type CreatePaymentInput struct {
	AccountID      uuid.UUID
	Amount         float64
	Currency       string
	Description    string
	IdempotencyKey string
}

// PaymentOutput serializa dados de pagamento para saída
type PaymentOutput struct {
	ID             string    `json:"id"`
	AccountID      string    `json:"accountId"`
	Amount         float64   `json:"amount"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"`
	Description    string    `json:"description,omitempty"`
	IdempotencyKey string    `json:"idempotencyKey"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func NewPaymentOutput(payment *entities.Payment) *PaymentOutput {
	return &PaymentOutput{
		ID:             payment.ID.String(),
		AccountID:      payment.AccountID.String(),
		Amount:         payment.Amount,
		Currency:       payment.Currency,
		Status:         string(payment.Status),
		Description:    payment.Description,
		IdempotencyKey: payment.IdempotencyKey,
		CreatedAt:      payment.CreatedAt,
		UpdatedAt:      payment.UpdatedAt,
	}
}

func NewPaymentOutputList(payments []*entities.Payment) []*PaymentOutput {
	outputs := make([]*PaymentOutput, len(payments))
	for i, payment := range payments {
		outputs[i] = NewPaymentOutput(payment)
	}
	return outputs
}

// CancelPaymentInput recebe dados para cancelar um pagamento
type CancelPaymentInput struct {
	PaymentID uuid.UUID
}
