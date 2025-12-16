// SPDX-License-Identifier: AGPL-3.0-or-later

package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/btouchard/shm/internal/store"
	"github.com/btouchard/shm/pkg/crypto"
	"github.com/btouchard/shm/pkg/logger"
)

// SignedRequestMiddleware vérifie que la requête est signée par l'instance
func SignedRequestMiddleware(db *store.Store, next http.HandlerFunc) http.HandlerFunc {
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

		pubKey, err := db.GetInstanceKey(instanceID)
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
