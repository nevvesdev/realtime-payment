package entities

import (
	"fmt"
)

type InvalidPaymentStatusTransitionError struct {
	CurrentStatus PaymentStatus
	TargetStatus  PaymentStatus
}

func (e *InvalidPaymentStatusTransitionError) Error() string {
	return fmt.Sprintf(
		"transição inválida de status de pagamento: %s -> %s",
		e.CurrentStatus,
		e.TargetStatus,
	)
}

func NewInvalidPaymentStatusTransitionError(current, target PaymentStatus) *InvalidPaymentStatusTransitionError {
	return &InvalidPaymentStatusTransitionError{
		CurrentStatus: current,
		TargetStatus:  target,
	}
}

type InvalidSettlementStatusTransitionError struct {
	CurrentStatus SettlementStatus
	TargetStatus  SettlementStatus
}

func (e *InvalidSettlementStatusTransitionError) Error() string {
	return fmt.Sprintf(
		"transição inválida de status de liquidação: %s -> %s",
		e.CurrentStatus,
		e.TargetStatus,
	)
}

func NewInvalidSettlementStatusTransitionError(current, target SettlementStatus) *InvalidSettlementStatusTransitionError {
	return &InvalidSettlementStatusTransitionError{
		CurrentStatus: current,
		TargetStatus:  target,
	}
}
