package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"br.com.nevvesdev/realtime-payment/internal/adapters/http"
	"br.com.nevvesdev/realtime-payment/internal/adapters/postgres"
	"br.com.nevvesdev/realtime-payment/internal/application/services"
	"br.com.nevvesdev/realtime-payment/internal/infrastructure/config"
	"br.com.nevvesdev/realtime-payment/internal/infrastructure/logger"
)

func main() {
	log := logger.Get()

	// Carregar configurações
	cfg, err := config.Load()
	if err != nil {
		log.WithField("error", err).Fatal("erro ao carregar configurações")
	}

	log.WithFields(map[string]interface{}{
		"env":  cfg.Server.Env,
		"port": cfg.Server.Port,
	}).Info("iniciando aplicação Real-time Payment Processing API")

	// Conectar ao banco de dados
	ctx := context.Background()
	dbConn, err := postgres.NewConnection(ctx, cfg.Database.URL)
	if err != nil {
		log.WithField("error", err).Fatal("erro ao conectar ao banco de dados")
	}
	defer dbConn.Close()

	log.Info("conexão ao banco de dados estabelecida com sucesso")

	// Inicializar repositórios
	paymentRepo := postgres.NewPostgresPaymentRepository(dbConn.Pool())
	settlementRepo := postgres.NewPostgresSettlementRepository(dbConn.Pool())
	paymentEventRepo := postgres.NewPostgresPaymentEventRepository(dbConn.Pool())
	outboxRepo := postgres.NewPostgresOutboxRepository(dbConn.Pool())

	// Inicializar serviços de aplicação
	paymentService := services.NewPaymentService(
		paymentRepo,
		settlementRepo,
		paymentEventRepo,
		outboxRepo,
	)

	settlementService := services.NewSettlementService(
		settlementRepo,
		paymentRepo,
		paymentEventRepo,
		outboxRepo,
	)

	log.Info("serviços de aplicação inicializados com sucesso")

	// Criar router HTTP
	router := http.NewRouter(
		paymentService,
		settlementService,
		paymentRepo,
		settlementRepo,
		paymentEventRepo,
	)

	// Canal para sinais de interrupção
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar servidor em goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- router.Serve(cfg.Server.Port)
	}()

	// Aguardar sinal de interrupção ou erro do servidor
	select {
	case <-sigChan:
		log.Info("recebido sinal de interrupção, encerrando gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), context.Background())
		defer cancel()
		// Aqui você poderia adicionar cleanup de recursos
		log.Info("aplicação encerrada com sucesso")
	case err := <-serverErr:
		if err != nil {
			log.WithField("error", err).Fatal("erro ao iniciar servidor")
		}
	}
}
