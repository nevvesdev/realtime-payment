package entities

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type PaymentEventType string

const (
	EventTypePaymentCreated      PaymentEventType = "PAYMENT_CREATED"
	EventTypePaymentProcessing   PaymentEventType = "PAYMENT_PROCESSING"
	EventTypePaymentCompleted    PaymentEventType = "PAYMENT_COMPLETED"
	EventTypePaymentFailed       PaymentEventType = "PAYMENT_FAILED"
	EventTypePaymentCancelled    PaymentEventType = "PAYMENT_CANCELLED"
	EventTypeSettlementCreated   PaymentEventType = "SETTLEMENT_CREATED"
	EventTypeSettlementCompleted PaymentEventType = "SETTLEMENT_COMPLETED"
	EventTypeSettlementFailed    PaymentEventType = "SETTLEMENT_FAILED"
)

type PaymentEvent struct {
	ID        uuid.UUID
	PaymentID uuid.UUID
	EventType PaymentEventType
	EventData json.RawMessage
	CreatedAt time.Time
}

func NewPaymentEvent(
	paymentID uuid.UUID,
	eventType PaymentEventType,
	eventData interface{},
) (*PaymentEvent, error) {
	data, err := json.Marshal(eventData)
	if err != nil {
		return nil, err
	}

	return &PaymentEvent{
		ID:        uuid.New(),
		PaymentID: paymentID,
		EventType: eventType,
		EventData: data,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (pe *PaymentEvent) UnmarshalData(v interface{}) error {
	return json.Unmarshal(pe.EventData, v)
}
