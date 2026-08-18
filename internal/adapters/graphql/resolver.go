package graphql

import (
	"br.com.nevvesdev/realtime-payment/internal/application/services"
	"br.com.nevvesdev/realtime-payment/internal/domain/repositories"
)

type Resolver struct {
	paymentService    *services.PaymentService
	settlementService *services.SettlementService
	paymentRepo       repositories.PaymentRepository
	settlementRepo    repositories.SettlementRepository
	paymentEventRepo  repositories.PaymentEventRepository
	subscriptionMgr   *SubscriptionManager
}

func NewResolver(
	paymentService *services.PaymentService,
	settlementService *services.SettlementService,
	paymentRepo repositories.PaymentRepository,
	settlementRepo repositories.SettlementRepository,
	paymentEventRepo repositories.PaymentEventRepository,
	subscriptionMgr *SubscriptionManager,
) *Resolver {
	return &Resolver{
		paymentService:    paymentService,
		settlementService: settlementService,
		paymentRepo:       paymentRepo,
		settlementRepo:    settlementRepo,
		paymentEventRepo:  paymentEventRepo,
		subscriptionMgr:   subscriptionMgr,
	}
}
