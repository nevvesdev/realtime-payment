package dto

import (
	"time"

	"br.com.nevvesdev/realtime-payment/internal/domain/entities"
)

// PaymentEventOutput serializa eventos de pagamento para saída
type PaymentEventOutput struct {
	ID        string    `json:"id"`
	PaymentID string    `json:"paymentId"`
	EventType string    `json:"eventType"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewPaymentEventOutput(event *entities.PaymentEvent) *PaymentEventOutput {
	return &PaymentEventOutput{
		ID:        event.ID.String(),
		PaymentID: event.PaymentID.String(),
		EventType: string(event.EventType),
		CreatedAt: event.CreatedAt,
	}
}

func NewPaymentEventOutputList(events []*entities.PaymentEvent) []*PaymentEventOutput {
	outputs := make([]*PaymentEventOutput, len(events))
	for i, event := range events {
		outputs[i] = NewPaymentEventOutput(event)
	}
	return outputs
}
