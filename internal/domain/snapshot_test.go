// SPDX-License-Identifier: AGPL-3.0-or-later

package domain

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestNewMetrics(t *testing.T) {
	tests := []struct {
		name    string
		input   json.RawMessage
		wantErr error
	}{
		{
			name:    "valid metrics",
			input:   json.RawMessage(`{"cpu": 0.5, "memory": 1024, "status": "ok"}`),
			wantErr: nil,
		},
		{
			name:    "empty input",
			input:   nil,
			wantErr: nil,
		},
		{
			name:    "empty object",
			input:   json.RawMessage(`{}`),
			wantErr: nil,
		},
		{
			name:    "invalid JSON",
			input:   json.RawMessage(`{invalid`),
			wantErr: ErrInvalidMetrics,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMetrics(tt.input)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if m == nil {
					t.Error("expected non-nil metrics")
				}
			}
		})
	}
}

func TestMetrics_GetFloat64(t *testing.T) {
	m := Metrics{
		"cpu":     0.5,
		"memory":  float64(1024),
		"status":  "ok",
		"enabled": true,
	}

	tests := []struct {
		key      string
		expected float64
		ok       bool
	}{
		{"cpu", 0.5, true},
		{"memory", 1024, true},
		{"status", 0, false},
		{"enabled", 0, false},
		{"missing", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			val, ok := m.GetFloat64(tt.key)
			if ok != tt.ok {
				t.Errorf("expected ok=%v, got ok=%v", tt.ok, ok)
			}
			if val != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, val)
			}
		})
	}
}

func TestMetrics_GetString(t *testing.T) {
	m := Metrics{
		"status": "ok",
		"cpu":    0.5,
	}

	tests := []struct {
		key      string
		expected string
		ok       bool
	}{
		{"status", "ok", true},
		{"cpu", "", false},
		{"missing", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			val, ok := m.GetString(tt.key)
			if ok != tt.ok {
				t.Errorf("expected ok=%v, got ok=%v", tt.ok, ok)
			}
			if val != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, val)
			}
		})
	}
}

func TestNewSnapshot(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	validMetrics := json.RawMessage(`{"cpu": 0.5}`)
	now := time.Now().UTC()

	tests := []struct {
		name       string
		instanceID string
		timestamp  time.Time
		metrics    json.RawMessage
		wantErr    error
	}{
		{
			name:       "valid snapshot",
			instanceID: validUUID,
			timestamp:  now,
			metrics:    validMetrics,
			wantErr:    nil,
		},
		{
			name:       "invalid instance ID",
			instanceID: "invalid",
			timestamp:  now,
			metrics:    validMetrics,
			wantErr:    ErrInvalidInstanceID,
		},
		{
			name:       "zero timestamp",
			instanceID: validUUID,
			timestamp:  time.Time{},
			metrics:    validMetrics,
			wantErr:    ErrInvalidSnapshot,
		},
		{
			name:       "future timestamp",
			instanceID: validUUID,
			timestamp:  now.Add(1 * time.Hour),
			metrics:    validMetrics,
			wantErr:    ErrInvalidSnapshot,
		},
		{
			name:       "invalid metrics JSON",
			instanceID: validUUID,
			timestamp:  now,
			metrics:    json.RawMessage(`{invalid`),
			wantErr:    ErrInvalidMetrics,
		},
		{
			name:       "empty metrics",
			instanceID: validUUID,
			timestamp:  now,
			metrics:    nil,
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snap, err := NewSnapshot(tt.instanceID, tt.timestamp, tt.metrics)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if snap == nil {
					t.Error("expected non-nil snapshot")
				}
			}
		})
	}
}

func TestSnapshot_Age(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	past := time.Now().UTC().Add(-1 * time.Hour)

	snap, err := NewSnapshot(validUUID, past, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	age := snap.Age()
	if age < 59*time.Minute || age > 61*time.Minute {
		t.Errorf("expected age ~1h, got %v", age)
	}
}
