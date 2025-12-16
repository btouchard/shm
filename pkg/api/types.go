// SPDX-License-Identifier: AGPL-3.0-or-later

package api

import (
	"encoding/json"
	"time"
)

// RegisterRequest est envoyé une seule fois à l'installation
type RegisterRequest struct {
	InstanceID     string `json:"instance_id"`
	PublicKey      string `json:"public_key"` // Hex encoded string
	AppName        string `json:"app_name"`
	AppVersion     string `json:"app_version"`
	DeploymentMode string `json:"deployment_mode"`
	Environment    string `json:"environment"`
	OSArch         string `json:"os_arch"`
}

// SnapshotRequest contient les métriques périodiques
type SnapshotRequest struct {
	InstanceID string          `json:"instance_id"`
	Timestamp  time.Time       `json:"timestamp"`
	Metrics    json.RawMessage `json:"metrics"` // Agnostique: {"docs": 10, "cpu": 0.5...}
}

// GenericResponse pour les retours API
type GenericResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// InstanceSummary objet retourné dans la liste
type InstanceSummary struct {
	InstanceID     string                 `json:"instance_id"`
	AppName        string                 `json:"app_name"`
	AppVersion     string                 `json:"app_version"`
	Environment    string                 `json:"environment"`
	Status         string                 `json:"status"`
	LastSeenAt     time.Time              `json:"last_seen_at"`
	DeploymentMode string                 `json:"deployment_mode"`
	Metrics        map[string]interface{} `json:"metrics"`
}

// DashboardStats compteurs globaux
type DashboardStats struct {
	TotalInstances  int              `json:"total_instances"`
	ActiveInstances int              `json:"active_instances"`
	GlobalMetrics   map[string]int64 `json:"global_metrics"`
}
