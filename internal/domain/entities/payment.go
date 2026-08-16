package entities

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "PENDING"
	PaymentStatusProcessing PaymentStatus = "PROCESSING"
	PaymentStatusCompleted  PaymentStatus = "COMPLETED"
	PaymentStatusFailed     PaymentStatus = "FAILED"
	PaymentStatusCancelled  PaymentStatus = "CANCELLED"
)

type Payment struct {
	ID             uuid.UUID
	AccountID      uuid.UUID
	Amount         float64
	Currency       string
	Status         PaymentStatus
	Description    string
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewPayment(
	accountID uuid.UUID,
	amount float64,
	currency string,
	description string,
	idempotencyKey string,
) *Payment {
	now := time.Now().UTC()

	return &Payment{
		ID:             uuid.New(),
		AccountID:      accountID,
		Amount:         amount,
		Currency:       currency,
		Status:         PaymentStatusPending,
		Description:    description,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (p *Payment) CanTransitionTo(newStatus PaymentStatus) bool {
	transitions := map[PaymentStatus][]PaymentStatus{
		PaymentStatusPending: {
			PaymentStatusProcessing,
			PaymentStatusCancelled,
			PaymentStatusFailed,
		},
		PaymentStatusProcessing: {
			PaymentStatusCompleted,
			PaymentStatusFailed,
		},
		PaymentStatusCompleted: {},
		PaymentStatusFailed:    {PaymentStatusPending},
		PaymentStatusCancelled: {},
	}

	allowedTransitions, exists := transitions[p.Status]
	if !exists {
		return false
	}

	for _, transition := range allowedTransitions {
		if transition == newStatus {
			return true
		}
	}

	return false
}

func (p *Payment) TransitionTo(newStatus PaymentStatus) error {
	if !p.CanTransitionTo(newStatus) {
		return NewInvalidPaymentStatusTransitionError(p.Status, newStatus)
	}

	p.Status = newStatus
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *Payment) IsIdempotencyKeyValid(key string) bool {
	return p.IdempotencyKey == key
}
