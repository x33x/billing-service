package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

var startTime = time.Now()

// there are no classes in Go, only structs
// this is my standard API response format
type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// w — answer to client, r — input request
// healthCheck returns service status and current timestamp
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data := map[string]any{
		"status":    "ok",
		"service":   "billing-service",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

// just ping service and uptime in seconds
func pingCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data := map[string]any{
		"service":        "billing-service",
		"version":        "0.1.0",
		"uptime_seconds": int64(time.Since(startTime).Seconds()),
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

func main() {
	// register handlers
	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/ping", pingCheck)

	log.Println("billing-service starting on :8080")

	// ListenAndServe blocks - server is working till stop
	// log.Fatal close app if server does not start
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
