package http

import (
	"net/http"
	"time"

	"br.com.nevvesdev/realtime-payment/internal/application/services"
	"br.com.nevvesdev/realtime-payment/internal/domain/repositories"
	"br.com.nevvesdev/realtime-payment/internal/infrastructure/logger"
)

type Router struct {
	mux *http.ServeMux
}

func NewRouter(
	paymentService *services.PaymentService,
	settlementService *services.SettlementService,
	paymentRepo repositories.PaymentRepository,
	settlementRepo repositories.SettlementRepository,
	paymentEventRepo repositories.PaymentEventRepository,
) *Router {
	log := logger.Get()

	mux := http.NewServeMux()

	// Criar servidor GraphQL
	graphqlServer := NewGraphQLServer(
		paymentService,
		settlementService,
		paymentRepo,
		settlementRepo,
		paymentEventRepo,
	)

	// Rota GraphQL
	mux.Handle("/graphql", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		graphqlServer.ServeHTTP(w, r)
	}))

	// Rota Playground
	mux.HandleFunc("/playground", graphqlServer.PlaygroundHandler())

	// Rota Health Check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Rota Readiness Check
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	log.Info("rotas HTTP configuradas com sucesso")

	return &Router{mux: mux}
}

func (r *Router) Serve(port string) error {
	log := logger.Get()

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.WithField("port", port).Info("iniciando servidor HTTP")

	return server.ListenAndServe()
}
