// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/btouchard/shm/internal/middleware"
	"github.com/btouchard/shm/internal/store"
	"github.com/btouchard/shm/pkg/api"
	"github.com/btouchard/shm/pkg/logger"
	"github.com/btouchard/shm/web"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
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

	http.HandleFunc("/v1/register", func(w http.ResponseWriter, r *http.Request) {
		logger.InfoCtx("HANDLER", "POST /v1/register")

		if r.Method != http.MethodPost {
			logger.WarnCtx("HANDLER", "Méthode non autorisée: %s", r.Method)
			http.Error(w, "Method not allowed", 405)
			return
		}
		var req api.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.ErrorCtx("HANDLER", "JSON invalide: %v", err)
			http.Error(w, "JSON invalide", 400)
			return
		}

		logger.InfoCtx("HANDLER", "Enregistrement instance: %s (app=%s, version=%s)", req.InstanceID, req.AppName, req.AppVersion)

		if err := db.RegisterInstance(req); err != nil {
			logger.ErrorCtx("HANDLER", "Erreur register pour %s: %v", req.InstanceID, err)
			http.Error(w, "Erreur serveur", 500)
			return
		}

		logger.InfoCtx("HANDLER", "Instance %s enregistrée avec succès", req.InstanceID)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(api.GenericResponse{Status: "ok", Message: "Registered"})
	})

	http.HandleFunc("/v1/activate", middleware.SignedRequestMiddleware(db, func(w http.ResponseWriter, r *http.Request) {
		instanceID := r.Header.Get("X-Instance-ID")
		logger.InfoCtx("HANDLER", "POST /v1/activate (instance=%s)", instanceID)

		if r.Method != http.MethodPost {
			logger.WarnCtx("HANDLER", "Méthode non autorisée: %s", r.Method)
			http.Error(w, "Method not allowed", 405)
			return
		}

		if err := db.ActivateInstance(instanceID); err != nil {
			logger.ErrorCtx("HANDLER", "Erreur activation %s: %v", instanceID, err)
			http.Error(w, "Erreur serveur", 500)
			return
		}

		logger.InfoCtx("HANDLER", "Instance %s activée avec succès", instanceID)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.GenericResponse{Status: "active", Message: "Instance activated successfully"})
	}))

	http.HandleFunc("/v1/snapshot", middleware.SignedRequestMiddleware(db, func(w http.ResponseWriter, r *http.Request) {
		instanceID := r.Header.Get("X-Instance-ID")
		logger.InfoCtx("HANDLER", "POST /v1/snapshot (instance=%s)", instanceID)

		if r.Method != http.MethodPost {
			logger.WarnCtx("HANDLER", "Méthode non autorisée: %s", r.Method)
			http.Error(w, "Method not allowed", 405)
			return
		}
		var req api.SnapshotRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.ErrorCtx("HANDLER", "JSON invalide: %v", err)
			http.Error(w, "JSON invalide", 400)
			return
		}

		if err := db.SaveSnapshot(req); err != nil {
			logger.ErrorCtx("HANDLER", "Erreur snapshot pour %s: %v", instanceID, err)
			http.Error(w, "Erreur serveur", 500)
			return
		}

		logger.InfoCtx("HANDLER", "Snapshot reçu pour instance %s", instanceID)
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(api.GenericResponse{Status: "ok", Message: "Snapshot received"})
	}))

	http.HandleFunc("/api/v1/admin/stats", func(w http.ResponseWriter, r *http.Request) {
		logger.InfoCtx("HANDLER", "GET /api/v1/admin/stats")

		stats, err := db.GetDashboardStats()
		if err != nil {
			logger.ErrorCtx("HANDLER", "Erreur récupération stats: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}

		logger.InfoCtx("HANDLER", "Stats récupérées avec succès (instances=%d)", stats.TotalInstances)
		_ = json.NewEncoder(w).Encode(stats)
	})

	http.HandleFunc("/api/v1/admin/instances", func(w http.ResponseWriter, r *http.Request) {
		logger.InfoCtx("HANDLER", "GET /api/v1/admin/instances")

		list, err := db.ListInstances(50) // Limit 50
		if err != nil {
			logger.ErrorCtx("HANDLER", "Erreur récupération instances: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}
		if list == nil {
			list = []api.InstanceSummary{}
		}

		logger.InfoCtx("HANDLER", "Liste instances récupérée (%d instances)", len(list))
		_ = json.NewEncoder(w).Encode(list)
	})

	http.HandleFunc("/api/v1/admin/metrics/", func(w http.ResponseWriter, r *http.Request) {
		// Extract app name from URL path: /api/v1/admin/metrics/{appName}
		appName := r.URL.Path[len("/api/v1/admin/metrics/"):]
		if appName == "" {
			http.Error(w, "App name required", 400)
			return
		}

		logger.InfoCtx("HANDLER", "GET /api/v1/admin/metrics/%s", appName)

		// Parse period parameter (default: 24h)
		periodParam := r.URL.Query().Get("period")
		periodHours := 24
		switch periodParam {
		case "7d":
			periodHours = 24 * 7
		case "30d":
			periodHours = 24 * 30
		case "3m":
			periodHours = 24 * 90
		case "1y":
			periodHours = 24 * 365
		case "all":
			periodHours = 24 * 365 * 10 // 10 years = effectively all
		default:
			periodHours = 24
		}

		data, err := db.GetMetricsTimeSeries(appName, periodHours)
		if err != nil {
			logger.ErrorCtx("HANDLER", "Erreur récupération métriques pour %s: %v", appName, err)
			http.Error(w, err.Error(), 500)
			return
		}

		logger.InfoCtx("HANDLER", "Métriques time series récupérées pour %s", appName)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(data)
	})

	http.Handle("/", http.FileServer(http.FS(web.Assets)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("═══════════════════════════════════════════════════")
	logger.InfoCtx("SERVER", "SHM (Self-Hosted Metrics) démarré")
	logger.InfoCtx("SERVER", "Port: %s", port)
	logger.InfoCtx("SERVER", "Endpoints: /v1/register, /v1/activate, /v1/snapshot, /api/v1/admin/*")
	logger.Info("═══════════════════════════════════════════════════")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
