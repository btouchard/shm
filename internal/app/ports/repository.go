// SPDX-License-Identifier: AGPL-3.0-or-later

// Package ports defines the interfaces (ports) used by the application layer.
// These interfaces are implemented by adapters (repositories, external services).
// Following hexagonal architecture: interfaces are declared where they are consumed.
package ports

import (
	"context"
	"time"

	"github.com/btouchard/shm/internal/domain"
)

// InstanceRepository defines persistence operations for instances.
type InstanceRepository interface {
	// Save persists an instance (insert or update).
	Save(ctx context.Context, instance *domain.Instance) error

	// FindByID retrieves an instance by its ID.
	// Returns domain.ErrInstanceNotFound if not found.
	FindByID(ctx context.Context, id domain.InstanceID) (*domain.Instance, error)

	// GetPublicKey retrieves the public key for an instance.
	// Returns domain.ErrInstanceNotFound if not found.
	// Returns domain.ErrInstanceRevoked if the instance is revoked.
	GetPublicKey(ctx context.Context, id domain.InstanceID) (domain.PublicKey, error)

	// UpdateStatus updates the status and last_seen_at timestamp.
	UpdateStatus(ctx context.Context, id domain.InstanceID, status domain.InstanceStatus) error
}

// SnapshotRepository defines persistence operations for snapshots.
type SnapshotRepository interface {
	// Save persists a snapshot and updates the instance heartbeat.
	Save(ctx context.Context, snapshot *domain.Snapshot) error

	// FindByInstanceID retrieves snapshots for an instance.
	FindByInstanceID(ctx context.Context, id domain.InstanceID, limit int) ([]*domain.Snapshot, error)

	// GetLatestByInstanceID retrieves the most recent snapshot for an instance.
	GetLatestByInstanceID(ctx context.Context, id domain.InstanceID) (*domain.Snapshot, error)
}

// DashboardStats holds aggregated statistics for the dashboard.
type DashboardStats struct {
	TotalInstances  int
	ActiveInstances int
	GlobalMetrics   map[string]int64
}

// InstanceSummary holds instance data with latest metrics for listing.
type InstanceSummary struct {
	ID             domain.InstanceID
	AppName        string
	AppVersion     string
	Environment    string
	Status         domain.InstanceStatus
	DeploymentMode string
	LastSeenAt     time.Time
	Metrics        domain.Metrics
}

// MetricsTimeSeries holds time-series data for charting.
type MetricsTimeSeries struct {
	Timestamps []time.Time
	Metrics    map[string][]float64
}

// DashboardReader defines read operations for the dashboard.
// Separated from write repositories for CQRS-lite pattern.
type DashboardReader interface {
	// GetStats returns aggregated dashboard statistics.
	GetStats(ctx context.Context) (DashboardStats, error)

	// ListInstances returns instances with their latest metrics.
	ListInstances(ctx context.Context, limit int) ([]InstanceSummary, error)

	// GetMetricsTimeSeries returns time-series metrics for an app.
	GetMetricsTimeSeries(ctx context.Context, appName string, since time.Time) (MetricsTimeSeries, error)
}
