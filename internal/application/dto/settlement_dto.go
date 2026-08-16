package dto

import (
	"time"

	"br.com.nevvesdev/realtime-payment/internal/domain/entities"
	"github.com/google/uuid"
)

// CreateSettlementInput recebe dados para criar uma liquidação
type CreateSettlementInput struct {
	PaymentID      uuid.UUID
	SettlementDate time.Time
}

// SettlementOutput serializa dados de liquidação para saída
type SettlementOutput struct {
	ID             string    `json:"id"`
	PaymentID      string    `json:"paymentId"`
	SettlementDate time.Time `json:"settlementDate"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func NewSettlementOutput(settlement *entities.Settlement) *SettlementOutput {
	return &SettlementOutput{
		ID:             settlement.ID.String(),
		PaymentID:      settlement.PaymentID.String(),
		SettlementDate: settlement.SettlementDate,
		Status:         string(settlement.Status),
		CreatedAt:      settlement.CreatedAt,
		UpdatedAt:      settlement.UpdatedAt,
	}
}

func NewSettlementOutputList(settlements []*entities.Settlement) []*SettlementOutput {
	outputs := make([]*SettlementOutput, len(settlements))
	for i, settlement := range settlements {
		outputs[i] = NewSettlementOutput(settlement)
	}
	return outputs
}

// UpdateSettlementStatusInput recebe dados para atualizar status
type UpdateSettlementStatusInput struct {
	SettlementID uuid.UUID
	NewStatus    entities.SettlementStatus
}
