// SPDX-License-Identifier: AGPL-3.0-or-later

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/btouchard/shm/internal/app/ports"
	"github.com/btouchard/shm/internal/domain"
)

// DashboardReader implements ports.DashboardReader for PostgreSQL.
type DashboardReader struct {
	db *sql.DB
}

// NewDashboardReader creates a new DashboardReader.
func NewDashboardReader(db *sql.DB) *DashboardReader {
	return &DashboardReader{db: db}
}

// GetStats returns aggregated dashboard statistics.
func (r *DashboardReader) GetStats(ctx context.Context) (ports.DashboardStats, error) {
	var stats ports.DashboardStats
	stats.GlobalMetrics = make(map[string]int64)

	// Get instance counts
	countsQuery := `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE last_seen_at > NOW() - INTERVAL '30 days')
		FROM instances
	`
	if err := r.db.QueryRowContext(ctx, countsQuery).Scan(&stats.TotalInstances, &stats.ActiveInstances); err != nil {
		return stats, fmt.Errorf("get instance counts: %w", err)
	}

	// Get aggregated metrics from latest snapshots
	metricsQuery := `
		SELECT data
		FROM (
			SELECT DISTINCT ON (instance_id) data
			FROM snapshots
			ORDER BY instance_id, snapshot_at DESC
		) as latest
	`
	rows, err := r.db.QueryContext(ctx, metricsQuery)
	if err != nil {
		return stats, fmt.Errorf("get latest metrics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rawJSON []byte
		if err := rows.Scan(&rawJSON); err != nil {
			continue
		}

		var metrics map[string]any
		if err := json.Unmarshal(rawJSON, &metrics); err != nil {
			continue
		}

		for key, val := range metrics {
			switch v := val.(type) {
			case float64:
				stats.GlobalMetrics[key] += int64(v)
			case int:
				stats.GlobalMetrics[key] += int64(v)
			}
		}
	}

	return stats, nil
}

// ListInstances returns instances with their latest metrics.
func (r *DashboardReader) ListInstances(ctx context.Context, limit int) ([]ports.InstanceSummary, error) {
	query := `
		SELECT
			i.instance_id, i.app_name, i.app_version, i.environment, i.status, i.last_seen_at, i.deployment_mode,
			COALESCE(s.data, '{}'::jsonb),
			a.app_slug, a.github_url, a.github_stars, a.logo_url
		FROM instances i
		LEFT JOIN applications a ON i.application_id = a.id
		LEFT JOIN LATERAL (
			SELECT data FROM snapshots
			WHERE instance_id = i.instance_id
			ORDER BY snapshot_at DESC
			LIMIT 1
		) s ON true
		ORDER BY i.last_seen_at DESC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}
	defer rows.Close()

	var list []ports.InstanceSummary
	for rows.Next() {
		var instanceID, status string
		var summary ports.InstanceSummary
		var rawMetrics []byte
		var appSlug, githubURL, logoURL sql.NullString
		var githubStars sql.NullInt64

		err := rows.Scan(
			&instanceID,
			&summary.AppName,
			&summary.AppVersion,
			&summary.Environment,
			&status,
			&summary.LastSeenAt,
			&summary.DeploymentMode,
			&rawMetrics,
			&appSlug,
			&githubURL,
			&githubStars,
			&logoURL,
		)
		if err != nil {
			return nil, fmt.Errorf("scan instance: %w", err)
		}

		summary.ID = domain.InstanceID(instanceID)
		summary.Status = domain.InstanceStatus(status)
		_ = json.Unmarshal(rawMetrics, &summary.Metrics)

		// Add application metadata
		if appSlug.Valid {
			summary.AppSlug = appSlug.String
		}
		if githubURL.Valid {
			summary.GitHubURL = githubURL.String
		}
		if githubStars.Valid {
			summary.GitHubStars = int(githubStars.Int64)
		}
		if logoURL.Valid {
			summary.LogoURL = logoURL.String
		}

		list = append(list, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate instances: %w", err)
	}

	return list, nil
}

// GetMetricsTimeSeries returns time-series metrics for an app.
func (r *DashboardReader) GetMetricsTimeSeries(ctx context.Context, appName string, since time.Time) (ports.MetricsTimeSeries, error) {
	query := `
		SELECT s.snapshot_at, s.data
		FROM snapshots s
		JOIN instances i ON s.instance_id = i.instance_id
		WHERE i.app_name = $1
		  AND s.snapshot_at > $2
		ORDER BY s.snapshot_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, appName, since)
	if err != nil {
		return ports.MetricsTimeSeries{}, fmt.Errorf("get metrics time series: %w", err)
	}
	defer rows.Close()

	// Aggregate metrics by timestamp
	timestampMap := make(map[time.Time]map[string]float64)
	var timestamps []time.Time

	for rows.Next() {
		var snapshotAt time.Time
		var rawMetrics []byte

		if err := rows.Scan(&snapshotAt, &rawMetrics); err != nil {
			continue
		}

		var metrics map[string]any
		if err := json.Unmarshal(rawMetrics, &metrics); err != nil {
			continue
		}

		if _, exists := timestampMap[snapshotAt]; !exists {
			timestampMap[snapshotAt] = make(map[string]float64)
			timestamps = append(timestamps, snapshotAt)
		}

		for key, val := range metrics {
			if v, ok := val.(float64); ok {
				timestampMap[snapshotAt][key] += v
			}
		}
	}

	// Build result
	result := ports.MetricsTimeSeries{
		Timestamps: timestamps,
		Metrics:    make(map[string][]float64),
	}

	for _, ts := range timestamps {
		for metricKey, value := range timestampMap[ts] {
			result.Metrics[metricKey] = append(result.Metrics[metricKey], value)
		}
	}

	return result, nil
}
