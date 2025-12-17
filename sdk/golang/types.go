// SPDX-License-Identifier: MIT

package golang

import (
	"encoding/json"
	"time"
)

// RegisterRequest is the payload for instance registration.
type RegisterRequest struct {
	InstanceID     string `json:"instance_id"`
	PublicKey      string `json:"public_key"`
	AppName        string `json:"app_name"`
	AppVersion     string `json:"app_version"`
	DeploymentMode string `json:"deployment_mode"`
	Environment    string `json:"environment"`
	OSArch         string `json:"os_arch"`
}

// SnapshotRequest is the payload for snapshot submission.
type SnapshotRequest struct {
	InstanceID string          `json:"instance_id"`
	Timestamp  time.Time       `json:"timestamp"`
	Metrics    json.RawMessage `json:"metrics"`
}
