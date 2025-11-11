package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/pkg/health"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// MetricsServer manages the HTTP server for metrics and health endpoints
type MetricsServer struct {
	server        *http.Server
	logger        *zap.Logger
	healthChecker *health.HealthChecker
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(port int, healthChecker *health.HealthChecker, logger *zap.Logger) *MetricsServer {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}

	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Health check endpoint
	if healthChecker != nil {
		mux.HandleFunc("/health", healthChecker.Handler())
	}

	// Root endpoint with service info
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Weather Service Metrics Server\n")
		fmt.Fprintf(w, "Endpoints:\n")
		fmt.Fprintf(w, "  /health  - Health check\n")
		fmt.Fprintf(w, "  /metrics - Prometheus metrics\n")
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &MetricsServer{
		server:        server,
		logger:        logger,
		healthChecker: healthChecker,
	}
}

// Start starts the metrics server
func (ms *MetricsServer) Start() error {
	ms.logger.Info("Starting metrics server",
		zap.String("address", ms.server.Addr))

	if err := ms.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("metrics server error: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the metrics server
func (ms *MetricsServer) Shutdown(ctx context.Context) error {
	ms.logger.Info("Shutting down metrics server")

	if err := ms.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("metrics server shutdown error: %w", err)
	}

	ms.logger.Info("Metrics server stopped")
	return nil
}
