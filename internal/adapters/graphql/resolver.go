package graphql

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"

	"br.com.nevvesdev/realtime-payment/internal/adapters/graphql/generated"
)

type Resolver struct{}

// CreatePayment is the resolver for the createPayment field.
func (r *mutationResolver) CreatePayment(ctx context.Context, input generated.CreatePaymentInput) (*generated.Payment, error) {
	panic("not implemented")
}

// CancelPayment is the resolver for the cancelPayment field.
func (r *mutationResolver) CancelPayment(ctx context.Context, id string) (*generated.Payment, error) {
	panic("not implemented")
}

// Payment is the resolver for the payment field.
func (r *queryResolver) Payment(ctx context.Context, id string) (*generated.Payment, error) {
	panic("not implemented")
}

// Payments is the resolver for the payments field.
func (r *queryResolver) Payments(ctx context.Context, accountID string, limit *int, offset *int) ([]*generated.Payment, error) {
	panic("not implemented")
}

// Settlement is the resolver for the settlement field.
func (r *queryResolver) Settlement(ctx context.Context, id string) (*generated.Settlement, error) {
	panic("not implemented")
}

// PaymentEvents is the resolver for the paymentEvents field.
func (r *queryResolver) PaymentEvents(ctx context.Context, paymentID string) ([]*generated.PaymentEvent, error) {
	panic("not implemented")
}

// PaymentStatusChanged is the resolver for the paymentStatusChanged field.
func (r *subscriptionResolver) PaymentStatusChanged(ctx context.Context, accountID string) (<-chan *generated.Payment, error) {
	panic("not implemented")
}

// SettlementStatusChanged is the resolver for the settlementStatusChanged field.
func (r *subscriptionResolver) SettlementStatusChanged(ctx context.Context, paymentID string) (<-chan *generated.Settlement, error) {
	panic("not implemented")
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type (
	mutationResolver     struct{ *Resolver }
	queryResolver        struct{ *Resolver }
	subscriptionResolver struct{ *Resolver }
)
