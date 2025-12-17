// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/btouchard/shm/internal/config"
	"github.com/btouchard/shm/internal/middleware"
	"github.com/btouchard/shm/internal/store"
	"github.com/btouchard/shm/pkg/api"
	"github.com/btouchard/shm/pkg/logger"
	"github.com/btouchard/shm/web"
)

func main() {
	dbURL := os.Getenv("SHM_DB_DSN")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/metrics?sslmode=disable"
	}

	logger.Info("Tentative de connexion à PostgreSQL...")
	db, err := store.NewStore(dbURL)
	if err != nil {
		logger.Error("Impossible de se connecter à la DB: %v", err)
		log.Fatalf("Impossible de se connecter à la DB: %v", err)
	}
	logger.InfoCtx("DATABASE", "Connecté à PostgreSQL avec succès")

	rlConfig := config.LoadRateLimitConfig()
	rl := middleware.NewRateLimiter(rlConfig)
	defer rl.Stop()

	if rlConfig.Enabled {
		logger.InfoCtx("RATELIMIT", "Rate limiting enabled")
	}

	router := api.NewRouter(db, rl)
	router.Handle("/", http.FileServer(http.FS(web.Assets)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("═══════════════════════════════════════════════════")
	logger.InfoCtx("SERVER", "SHM (Self-Hosted Metrics) démarré")
	logger.InfoCtx("SERVER", "Port: %s", port)
	logger.InfoCtx("SERVER", "Endpoints: /v1/register, /v1/activate, /v1/snapshot, /api/v1/admin/*")
	logger.Info("═══════════════════════════════════════════════════")

	log.Fatal(http.ListenAndServe(":"+port, router))
}
