package errors

import "fmt"

type ErrorCode string

const (
	ErrInvalidInput         ErrorCode = "INVALID_INPUT"
	ErrPaymentNotFound      ErrorCode = "PAYMENT_NOT_FOUND"
	ErrPaymentAlreadyExists ErrorCode = "PAYMENT_ALREADY_EXISTS"
	ErrInvalidPaymentStatus ErrorCode = "INVALID_PAYMENT_STATUS"
	ErrSettlementNotFound   ErrorCode = "SETTLEMENT_NOT_FOUND"
	ErrDatabaseError        ErrorCode = "DATABASE_ERROR"
	ErrKafkaError           ErrorCode = "KAFKA_ERROR"
	ErrInternalError        ErrorCode = "INTERNAL_ERROR"
)

type DomainError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewDomainError(code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

func NewDomainErrorWithCause(code ErrorCode, message string, cause error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

func IsErrorCode(err error, code ErrorCode) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == code
	}
	return false
}
