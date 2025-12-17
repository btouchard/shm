// SPDX-License-Identifier: AGPL-3.0-or-later

package api

import (
	"bytes"
	"io"
	"net/http"

	"github.com/btouchard/shm/internal/middleware"
	"github.com/btouchard/shm/pkg/crypto"
	"github.com/btouchard/shm/pkg/logger"
)

func NewRouter(store Store, rl *middleware.RateLimiter) *http.ServeMux {
	h := NewHandlers(store)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/healthcheck", h.Healthcheck)
	mux.HandleFunc("/v1/register", rl.RegisterMiddleware(h.Register))
	mux.HandleFunc("/v1/activate", rl.RegisterMiddleware(signedRequestMiddleware(store, h.Activate)))
	mux.HandleFunc("/v1/snapshot", rl.SnapshotMiddleware(signedRequestMiddleware(store, h.Snapshot)))
	mux.HandleFunc("/api/v1/admin/stats", rl.AdminMiddleware(h.AdminStats))
	mux.HandleFunc("/api/v1/admin/instances", rl.AdminMiddleware(h.AdminInstances))
	mux.HandleFunc("/api/v1/admin/metrics/", rl.AdminMiddleware(h.AdminMetrics))
	return mux
}

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
