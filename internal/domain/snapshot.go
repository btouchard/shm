// SPDX-License-Identifier: AGPL-3.0-or-later

package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// Metrics represents schema-agnostic telemetry data.
// Stored as JSON, can contain any numeric or string values.
type Metrics map[string]any

// NewMetrics creates Metrics from raw JSON bytes.
func NewMetrics(raw json.RawMessage) (Metrics, error) {
	if len(raw) == 0 {
		return make(Metrics), nil
	}

	var m Metrics
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidMetrics, err)
	}
	return m, nil
}

// Raw returns the JSON representation of the metrics.
func (m Metrics) Raw() (json.RawMessage, error) {
	return json.Marshal(m)
}

// GetFloat64 retrieves a numeric metric value.
// Returns 0 and false if the key doesn't exist or isn't numeric.
func (m Metrics) GetFloat64(key string) (float64, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}

// GetString retrieves a string metric value.
func (m Metrics) GetString(key string) (string, bool) {
	v, ok := m[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// SumNumeric returns the sum of all numeric values in the metrics.
func (m Metrics) SumNumeric() map[string]float64 {
	result := make(map[string]float64)
	for key, val := range m {
		if v, ok := m.GetFloat64(key); ok {
			result[key] = v
		} else {
			_ = val // ignore non-numeric
		}
	}
	return result
}

// Snapshot represents a point-in-time telemetry capture from an instance.
type Snapshot struct {
	ID         int64
	InstanceID InstanceID
	SnapshotAt time.Time
	Metrics    Metrics
}

// NewSnapshot creates a new Snapshot with validation.
func NewSnapshot(instanceID string, timestamp time.Time, metrics json.RawMessage) (*Snapshot, error) {
	id, err := NewInstanceID(instanceID)
	if err != nil {
		return nil, err
	}

	if timestamp.IsZero() {
		return nil, fmt.Errorf("%w: timestamp is required", ErrInvalidSnapshot)
	}

	// Normalize to UTC
	timestamp = timestamp.UTC()

	// Reject future timestamps (with small tolerance for clock skew)
	if timestamp.After(time.Now().UTC().Add(5 * time.Minute)) {
		return nil, fmt.Errorf("%w: timestamp is in the future", ErrInvalidSnapshot)
	}

	m, err := NewMetrics(metrics)
	if err != nil {
		return nil, err
	}

	return &Snapshot{
		InstanceID: id,
		SnapshotAt: timestamp,
		Metrics:    m,
	}, nil
}

// Age returns how old the snapshot is.
func (s *Snapshot) Age() time.Duration {
	return time.Since(s.SnapshotAt)
}
