package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"gopherpay/internal/api"
	"gopherpay/internal/config"
	"gopherpay/internal/db"
	"gopherpay/internal/logger"
	"gopherpay/internal/wallet"
	"gopherpay/internal/worker"
)

func main() {

	// =====================================
	// Setup structured logger
	// =====================================
	log := logger.New()
	slog.SetDefault(log)

	ctx := context.Background()

	// =====================================
	// Load configuration
	// =====================================
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// =====================================
	// Connect to database
	// =====================================
	database, err := db.NewPostgresConnection(ctx, cfg)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	// =====================================
	// Initialize wallet service
	// =====================================
	repo := wallet.NewPostgresRepository(database)
	service := wallet.NewWalletService(database, repo)

	// =====================================
	// Start worker pool
	// =====================================
	pool := worker.NewWorkerPool(service, cfg.WorkerPoolSize)
	pool.Start(ctx, cfg.WorkerCount)

	// =====================================
	// Setup HTTP server
	// =====================================
	handler := &api.Handler{
		Pool: pool,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/transfer", handler.Transfer)

	server := http.Server{
		Addr:    cfg.ServerHost + ":" + cfg.ServerPort,
		Handler: api.RequestIDMiddleware(mux),
	}

	slog.Info("server started",
		"host", cfg.ServerHost,
		"port", cfg.ServerPort,
		"workers", cfg.WorkerCount,
	)

	// =====================================
	// Start server
	// =====================================
	if err := server.ListenAndServe(); err != nil {
		slog.Error("server crashed", "error", err)
		os.Exit(1)
	}
}
