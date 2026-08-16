package entities

import (
	"time"

	"github.com/google/uuid"
)

type SettlementStatus string

const (
	SettlementStatusPending SettlementStatus = "PENDING"
	SettlementStatusSettled SettlementStatus = "SETTLED"
	SettlementStatusFailed  SettlementStatus = "FAILED"
)

type Settlement struct {
	ID             uuid.UUID
	PaymentID      uuid.UUID
	SettlementDate time.Time
	Status         SettlementStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewSettlement(paymentID uuid.UUID, settlementDate time.Time) *Settlement {
	now := time.Now().UTC()

	return &Settlement{
		ID:             uuid.New(),
		PaymentID:      paymentID,
		SettlementDate: settlementDate,
		Status:         SettlementStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (s *Settlement) CanTransitionTo(newStatus SettlementStatus) bool {
	transitions := map[SettlementStatus][]SettlementStatus{
		SettlementStatusPending: {
			SettlementStatusSettled,
			SettlementStatusFailed,
		},
		SettlementStatusSettled: {},
		SettlementStatusFailed:  {SettlementStatusPending},
	}

	allowedTransitions, exists := transitions[s.Status]
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

func (s *Settlement) TransitionTo(newStatus SettlementStatus) error {
	if !s.CanTransitionTo(newStatus) {
		return NewInvalidSettlementStatusTransitionError(s.Status, newStatus)
	}

	s.Status = newStatus
	s.UpdatedAt = time.Now().UTC()
	return nil
}
