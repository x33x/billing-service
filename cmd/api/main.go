package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/x33x/billing-service/internal/config"
	"github.com/x33x/billing-service/internal/db"
	"github.com/x33x/billing-service/internal/domain"
	"github.com/x33x/billing-service/internal/handler"
	"github.com/x33x/billing-service/internal/logger"
	"github.com/x33x/billing-service/internal/repository"
	"github.com/x33x/billing-service/internal/service"
)

var startTime = time.Now()

func main() {
	ctx := context.Background()

	// load .env
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found")
	}
	logger.Setup(os.Getenv("LOG_LEVEL"))

	cfg, err := config.Load()

	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	// connect to db
	database, err := db.New(ctx, cfg.DSN)
	if err != nil {
		slog.Error("connect to db", "error", err)
		os.Exit(1)
	}
	defer database.Close()
	slog.Info("connected to database")

	// dependency injection - collect layers from down to up
	accountRepo := repository.NewAccountRepository(database)
	ledgerRepo := repository.NewLedgerRepository(database)
	txRepo := repository.NewTransactionRepository(database, ledgerRepo)
	feeConfig := domain.FeeConfig{Rate: cfg.FeeRate}
	paymentSvc := service.NewPaymentService(accountRepo, txRepo, feeConfig)
	paymentHandler := handler.NewPaymentHandler(paymentSvc)

	// routing
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthCheck)
	mux.HandleFunc("GET /ping", pingCheck)
	mux.HandleFunc("POST /payments", paymentHandler.CreatePayment)
	mux.HandleFunc("GET /accounts/{id}/balance", paymentHandler.GetBalance)
	mux.HandleFunc("GET /accounts/{id}/transactions", paymentHandler.GetTransactions)

	slog.Info("server starting", "addr", cfg.ServerAddr)

	// ListenAndServe blocks - server is working till stop
	// log.Fatal close app if server does not start
	if err := http.ListenAndServe(cfg.ServerAddr, mux); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

// healthCheck returns service status and current timestamp
func healthCheck(w http.ResponseWriter, r *http.Request) {
	handler.WriteJSON(w, http.StatusOK, handler.APIResponse{
		Success: true,
		Data: map[string]any{
			"status":    "ok",
			"service":   "billing-service",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// just ping service and uptime in seconds
func pingCheck(w http.ResponseWriter, r *http.Request) {
	handler.WriteJSON(w, http.StatusOK, handler.APIResponse{
		Success: true,
		Data: map[string]any{
			"service":        "billing-service",
			"version":        "0.1.0",
			"uptime_seconds": int64(time.Since(startTime).Seconds()),
			"timestamp":      time.Now().UTC().Format(time.RFC3339),
		},
	})
}
