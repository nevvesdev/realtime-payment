package http

import (
	"net/http"

	"br.com.nevvesdev/realtime-payment/internal/adapters/graphql"
	"br.com.nevvesdev/realtime-payment/internal/adapters/graphql/generated"
	"br.com.nevvesdev/realtime-payment/internal/application/services"
	"br.com.nevvesdev/realtime-payment/internal/domain/repositories"
	"br.com.nevvesdev/realtime-payment/internal/infrastructure/logger"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

type GraphQLServer struct {
	handler http.Handler
}

func NewGraphQLServer(
	paymentService *services.PaymentService,
	settlementService *services.SettlementService,
	paymentRepo repositories.PaymentRepository,
	settlementRepo repositories.SettlementRepository,
	paymentEventRepo repositories.PaymentEventRepository,
) *GraphQLServer {
	log := logger.Get()

	// Criar gerenciador de subscriptions
	subscriptionMgr := graphql.NewSubscriptionManager()

	// Criar resolver
	resolver := graphql.NewResolver(
		paymentService,
		settlementService,
		paymentRepo,
		settlementRepo,
		paymentEventRepo,
		subscriptionMgr,
	)

	// Criar schema GraphQL
	schema := generated.NewExecutableSchema(generated.Config{
		Resolvers: resolver,
	})

	// Criar handler GraphQL com middleware
	gqlHandler := handler.NewDefaultServer(schema)

	// Middleware para logging
	gqlHandler.Use(&loggingMiddleware{log: log})

	// Middleware para recuperação de panics
	gqlHandler.Use(&panicRecoveryMiddleware{log: log})

	return &GraphQLServer{
		handler: gqlHandler,
	}
}

func (s *GraphQLServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *GraphQLServer) PlaygroundHandler() http.HandlerFunc {
	return playground.Handler("GraphQL", "/graphql")
}

// loggingMiddleware registra requisições GraphQL
type loggingMiddleware struct {
	log *logger.Logger
}

func (m *loggingMiddleware) ExtensionName() string {
	return "loggingMiddleware"
}

func (m *loggingMiddleware) Validate(schema interface{}) error {
	return nil
}

// panicRecoveryMiddleware recupera de panics em resolvers
type panicRecoveryMiddleware struct {
	log *logger.Logger
}

func (m *panicRecoveryMiddleware) ExtensionName() string {
	return "panicRecoveryMiddleware"
}

func (m *panicRecoveryMiddleware) Validate(schema interface{}) error {
	return nil
}
