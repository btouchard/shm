// SPDX-License-Identifier: AGPL-3.0-or-later

package api

import (
	"encoding/json"
	"net/http"

	"github.com/btouchard/shm/pkg/logger"
)

type Store interface {
	RegisterInstance(req RegisterRequest) error
	ActivateInstance(instanceID string) error
	SaveSnapshot(req SnapshotRequest) error
	GetInstanceKey(instanceID string) (string, error)
	GetDashboardStats() (DashboardStats, error)
	ListInstances(limit int) ([]InstanceSummary, error)
	GetMetricsTimeSeries(appName string, periodHours int) (map[string]interface{}, error)
}

type Handlers struct {
	store Store
}

func NewHandlers(store Store) *Handlers {
	return &Handlers{store: store}
}

func (h *Handlers) Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	logger.InfoCtx("HANDLER", "POST /v1/register")

	if r.Method != http.MethodPost {
		logger.WarnCtx("HANDLER", "Méthode non autorisée: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.ErrorCtx("HANDLER", "JSON invalide: %v", err)
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	logger.InfoCtx("HANDLER", "Enregistrement instance: %s (app=%s, version=%s)", req.InstanceID, req.AppName, req.AppVersion)

	if err := h.store.RegisterInstance(req); err != nil {
		logger.ErrorCtx("HANDLER", "Erreur register pour %s: %v", req.InstanceID, err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	logger.InfoCtx("HANDLER", "Instance %s enregistrée avec succès", req.InstanceID)
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(GenericResponse{Status: "ok", Message: "Registered"})
}

func (h *Handlers) Activate(w http.ResponseWriter, r *http.Request) {
	instanceID := r.Header.Get("X-Instance-ID")
	logger.InfoCtx("HANDLER", "POST /v1/activate (instance=%s)", instanceID)

	if r.Method != http.MethodPost {
		logger.WarnCtx("HANDLER", "Méthode non autorisée: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.store.ActivateInstance(instanceID); err != nil {
		logger.ErrorCtx("HANDLER", "Erreur activation %s: %v", instanceID, err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	logger.InfoCtx("HANDLER", "Instance %s activée avec succès", instanceID)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(GenericResponse{Status: "active", Message: "Instance activated successfully"})
}

func (h *Handlers) Snapshot(w http.ResponseWriter, r *http.Request) {
	instanceID := r.Header.Get("X-Instance-ID")
	logger.InfoCtx("HANDLER", "POST /v1/snapshot (instance=%s)", instanceID)

	if r.Method != http.MethodPost {
		logger.WarnCtx("HANDLER", "Méthode non autorisée: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.ErrorCtx("HANDLER", "JSON invalide: %v", err)
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	if err := h.store.SaveSnapshot(req); err != nil {
		logger.ErrorCtx("HANDLER", "Erreur snapshot pour %s: %v", instanceID, err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	logger.InfoCtx("HANDLER", "Snapshot reçu pour instance %s", instanceID)
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(GenericResponse{Status: "ok", Message: "Snapshot received"})
}

func (h *Handlers) AdminStats(w http.ResponseWriter, r *http.Request) {
	logger.InfoCtx("HANDLER", "GET /api/v1/admin/stats")

	stats, err := h.store.GetDashboardStats()
	if err != nil {
		logger.ErrorCtx("HANDLER", "Erreur récupération stats: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.InfoCtx("HANDLER", "Stats récupérées avec succès (instances=%d)", stats.TotalInstances)
	_ = json.NewEncoder(w).Encode(stats)
}

func (h *Handlers) AdminInstances(w http.ResponseWriter, r *http.Request) {
	logger.InfoCtx("HANDLER", "GET /api/v1/admin/instances")

	list, err := h.store.ListInstances(50)
	if err != nil {
		logger.ErrorCtx("HANDLER", "Erreur récupération instances: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []InstanceSummary{}
	}

	logger.InfoCtx("HANDLER", "Liste instances récupérée (%d instances)", len(list))
	_ = json.NewEncoder(w).Encode(list)
}

func (h *Handlers) AdminMetrics(w http.ResponseWriter, r *http.Request) {
	appName := r.URL.Path[len("/api/v1/admin/metrics/"):]
	if appName == "" {
		http.Error(w, "App name required", http.StatusBadRequest)
		return
	}

	logger.InfoCtx("HANDLER", "GET /api/v1/admin/metrics/%s", appName)

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

	data, err := h.store.GetMetricsTimeSeries(appName, periodHours)
	if err != nil {
		logger.ErrorCtx("HANDLER", "Erreur récupération métriques pour %s: %v", appName, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.InfoCtx("HANDLER", "Métriques time series récupérées pour %s", appName)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}
