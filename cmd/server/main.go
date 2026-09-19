package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/alert"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/config"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/controller"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/database"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/repository"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/router"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/service"
	ws "github.com/ericanthonywu/maremereso-olga/backend/internal/websocket"
)

func main() {
	// Structured logging
	// Debug logging prints request bodies and internal state; keep it out of
	// production, where the logs would accumulate customer phone numbers.
	logLevel := slog.LevelDebug
	if strings.EqualFold(os.Getenv("APP_ENV"), "production") {
		logLevel = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))

	slog.Info("Starting Maremereso Olga Backend Server...")

	// 1. Load Config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "err", err)
		os.Exit(1)
	}

	// 2. Connect Database
	dbPool, err := database.NewPool(cfg)
	if err != nil {
		slog.Error("Failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// 3. Initialize WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// 4. Initialize Dependency Layers (Separation of Concern)
	// Repository -> Service -> Controller -> Router
	alertSvc := alert.NewAlertService(cfg)
	repo := repository.NewRepository(dbPool)
	repo.EnsureFeedbackSchema(context.Background())
	repo.EnsureStaffCredentialsSchema(context.Background())
	svc := service.NewService(repo, cfg, hub)
	ctrl := controller.NewController(svc, hub, cfg)
	r := router.NewRouter(cfg, ctrl, alertSvc)

	// 5. HTTP Server setup with graceful shutdown
	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Port),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		// WriteTimeout must stay unset: it also caps WebSocket connections,
		// and a 15s cap was silently killing every live order feed.
		IdleTimeout: 120 * time.Second,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("HTTP server listening", "port", cfg.Port, "url", fmt.Sprintf("http://localhost:%s", cfg.Port))
		serverErrors <- server.ListenAndServe()
	}()

	// 6. Graceful Shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server error", "err", err)
		}

	case sig := <-shutdown:
		slog.Info("Shutdown signal received", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Graceful shutdown failed, forcing close", "err", err)
			_ = server.Close()
		}
		slog.Info("Server stopped cleanly")
	}
}
