// SPDX-License-Identifier: AGPL-3.0-or-later

package api

import (
	"bytes"
	"io"
	"net/http"

	"github.com/btouchard/shm/pkg/crypto"
	"github.com/btouchard/shm/pkg/logger"
)

// NewRouter creates and configures the HTTP router with all routes
func NewRouter(store Store) *http.ServeMux {
	mux := http.NewServeMux()
	h := NewHandlers(store)

	// SDK endpoints
	mux.HandleFunc("/v1/register", h.Register)
	mux.HandleFunc("/v1/activate", signedRequestMiddleware(store, h.Activate))
	mux.HandleFunc("/v1/snapshot", signedRequestMiddleware(store, h.Snapshot))

	// Admin endpoints (dashboard)
	mux.HandleFunc("/api/v1/admin/stats", h.AdminStats)
	mux.HandleFunc("/api/v1/admin/instances", h.AdminInstances)
	mux.HandleFunc("/api/v1/admin/metrics/", h.AdminMetrics)

	return mux
}

// signedRequestMiddleware verifies Ed25519 signatures on requests
func signedRequestMiddleware(store Store, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		instanceID := r.Header.Get("X-Instance-ID")
		signature := r.Header.Get("X-Signature")

		logger.DebugCtx("AUTH", "Tentative d'authentification pour instance: %s", instanceID)

		if instanceID == "" || signature == "" {
			logger.WarnCtx("AUTH", "Headers d'authentification manquants (instance=%s, signature présente=%v)", instanceID, signature != "")
			http.Error(w, "Headers d'authentification manquants", http.StatusUnauthorized)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logger.ErrorCtx("AUTH", "Erreur lecture body pour %s: %v", instanceID, err)
			http.Error(w, "Erreur lecture body", http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		pubKey, err := store.GetInstanceKey(instanceID)
		if err != nil {
			logger.ErrorCtx("AUTH", "Échec récupération clé pour %s: %v", instanceID, err)
			http.Error(w, "Non autorisé", http.StatusForbidden)
			return
		}

		if !crypto.Verify(pubKey, bodyBytes, signature) {
			logger.ErrorCtx("AUTH", "Signature invalide pour %s", instanceID)
			http.Error(w, "Signature invalide", http.StatusForbidden)
			return
		}

		logger.InfoCtx("AUTH", "Authentification réussie pour %s", instanceID)
		next(w, r)
	}
}
