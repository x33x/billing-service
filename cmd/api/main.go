package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/x33x/billing-service/internal/db"
	"github.com/x33x/billing-service/internal/handler"
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

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	// connect to db
	database, err := db.New(ctx, connString)
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer database.Close()
	log.Println("connected to database")

	// dependency injection - collect layers from down to up
	accountRepo := repository.NewAccountRepository(database)
	ledgerRepo := repository.NewLedgerRepository(database)
	txRepo := repository.NewTransactionRepository(database, ledgerRepo)
	paymentSvc := service.NewPaymentService(accountRepo, txRepo)
	paymentHandler := handler.NewPaymentHandler(paymentSvc)

	// routing
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthCheck)
	mux.HandleFunc("GET /ping", pingCheck)
	mux.HandleFunc("POST /payments", paymentHandler.CreatePayment)
	mux.HandleFunc("GET /accounts/{id}/balance", paymentHandler.GetBalance)
	mux.HandleFunc("GET /accounts/{id}/transactions", paymentHandler.GetTransactions)

	log.Println("billing-service starting on :8080")

	// ListenAndServe blocks - server is working till stop
	// log.Fatal close app if server does not start
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
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
