// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"log/slog"
	"net/http"

	"github.com/btouchard/shm/internal/adapters/postgres"
	"github.com/btouchard/shm/internal/app"
	"github.com/btouchard/shm/internal/middleware"
)

// RouterConfig holds the configuration for creating a new router.
type RouterConfig struct {
	Store       *postgres.Store
	RateLimiter *middleware.RateLimiter
	Logger      *slog.Logger
}

// NewRouter creates a fully wired HTTP router with all handlers and middleware.
func NewRouter(cfg RouterConfig) *http.ServeMux {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	// Create repositories from store
	instanceRepo := cfg.Store.InstanceRepository()
	snapshotRepo := cfg.Store.SnapshotRepository()
	dashboardReader := cfg.Store.DashboardReader()

	// Create application services
	instanceSvc := app.NewInstanceService(instanceRepo)
	snapshotSvc := app.NewSnapshotService(snapshotRepo, instanceRepo)
	dashboardSvc := app.NewDashboardService(dashboardReader)

	// Create HTTP handlers
	handlers := NewHandlers(instanceSvc, snapshotSvc, dashboardSvc, logger)

	// Create auth middleware
	authMW := NewAuthMiddlewareFromService(instanceSvc, logger)

	// Create router
	mux := http.NewServeMux()

	// Health check (no auth, no rate limit)
	mux.HandleFunc("/api/v1/healthcheck", handlers.Healthcheck)

	// Client endpoints
	rl := cfg.RateLimiter
	if rl != nil {
		mux.HandleFunc("/v1/register", rl.RegisterMiddleware(handlers.Register))
		mux.HandleFunc("/v1/activate", rl.RegisterMiddleware(authMW.RequireSignature(handlers.Activate)))
		mux.HandleFunc("/v1/snapshot", rl.SnapshotMiddleware(authMW.RequireSignature(handlers.Snapshot)))

		// Admin endpoints
		mux.HandleFunc("/api/v1/admin/stats", rl.AdminMiddleware(handlers.AdminStats))
		mux.HandleFunc("/api/v1/admin/instances", rl.AdminMiddleware(handlers.AdminInstances))
		mux.HandleFunc("/api/v1/admin/metrics/", rl.AdminMiddleware(handlers.AdminMetrics))
	} else {
		// No rate limiting (for testing)
		mux.HandleFunc("/v1/register", handlers.Register)
		mux.HandleFunc("/v1/activate", authMW.RequireSignature(handlers.Activate))
		mux.HandleFunc("/v1/snapshot", authMW.RequireSignature(handlers.Snapshot))
		mux.HandleFunc("/api/v1/admin/stats", handlers.AdminStats)
		mux.HandleFunc("/api/v1/admin/instances", handlers.AdminInstances)
		mux.HandleFunc("/api/v1/admin/metrics/", handlers.AdminMetrics)
	}

	return mux
}
