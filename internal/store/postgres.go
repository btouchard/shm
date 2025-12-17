// SPDX-License-Identifier: AGPL-3.0-or-later

package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/btouchard/shm/pkg/api"
	"github.com/btouchard/shm/pkg/logger"
	_ "github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

func NewStore(connStr string) (*Store, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) RegisterInstance(req api.RegisterRequest) error {
	logger.DebugCtx("STORE", "RegisterInstance: %s (app=%s)", req.InstanceID, req.AppName)

	query := `
		INSERT INTO instances (instance_id, public_key, app_name, app_version, deployment_mode, environment, os_arch, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (instance_id) DO UPDATE
		SET app_name = EXCLUDED.app_name, -- Au cas où le nom change
			app_version = EXCLUDED.app_version,
			last_seen_at = NOW();
	`
	_, err := s.db.Exec(query, req.InstanceID, req.PublicKey, req.AppName, req.AppVersion, req.DeploymentMode, req.Environment, req.OSArch)
	if err != nil {
		logger.ErrorCtx("STORE", "Erreur SQL RegisterInstance pour %s: %v", req.InstanceID, err)
		return err
	}

	logger.InfoCtx("STORE", "Instance %s enregistrée/mise à jour en DB", req.InstanceID)
	return nil
}

func (s *Store) ActivateInstance(instanceID string) error {
	logger.DebugCtx("STORE", "ActivateInstance: %s", instanceID)

	result, err := s.db.Exec(`UPDATE instances SET status = 'active', last_seen_at = NOW() WHERE instance_id = $1`, instanceID)
	if err != nil {
		logger.ErrorCtx("STORE", "Erreur SQL ActivateInstance pour %s: %v", instanceID, err)
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		logger.WarnCtx("STORE", "Instance %s non trouvée lors de l'activation", instanceID)
		return fmt.Errorf("instance not found")
	}

	logger.InfoCtx("STORE", "Instance %s activée en DB", instanceID)
	return nil
}

func (s *Store) GetInstanceKey(instanceID string) (string, error) {
	logger.DebugCtx("STORE", "GetInstanceKey: %s", instanceID)

	var pubKey string
	query := `SELECT public_key FROM instances WHERE instance_id = $1 AND status != 'revoked'`
	err := s.db.QueryRow(query, instanceID).Scan(&pubKey)
	if errors.Is(err, sql.ErrNoRows) {
		logger.WarnCtx("STORE", "Instance %s non trouvée ou révoquée", instanceID)
		return "", fmt.Errorf("instance non trouvée ou révoquée")
	}
	if err != nil {
		logger.ErrorCtx("STORE", "Erreur SQL GetInstanceKey pour %s: %v", instanceID, err)
		return "", err
	}

	logger.DebugCtx("STORE", "Clé publique récupérée pour %s", instanceID)
	return pubKey, nil
}

func (s *Store) SaveSnapshot(req api.SnapshotRequest) error {
	logger.DebugCtx("STORE", "SaveSnapshot: %s (timestamp=%s)", req.InstanceID, req.Timestamp)

	tx, err := s.db.Begin()
	if err != nil {
		logger.ErrorCtx("STORE", "Erreur démarrage transaction pour %s: %v", req.InstanceID, err)
		return err
	}

	metricsJSON, _ := json.Marshal(req.Metrics)
	_, err = tx.Exec(`INSERT INTO snapshots (instance_id, snapshot_at, data) VALUES ($1, $2, $3)`,
		req.InstanceID, req.Timestamp, metricsJSON)
	if err != nil {
		logger.ErrorCtx("STORE", "Erreur insertion snapshot pour %s: %v", req.InstanceID, err)
		_ = tx.Rollback()
		return err
	}

	_, err = tx.Exec(`UPDATE instances SET last_seen_at = NOW() WHERE instance_id = $1`, req.InstanceID)
	if err != nil {
		logger.ErrorCtx("STORE", "Erreur update heartbeat pour %s: %v", req.InstanceID, err)
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.ErrorCtx("STORE", "Erreur commit transaction pour %s: %v", req.InstanceID, err)
		return err
	}

	logger.InfoCtx("STORE", "Snapshot sauvegardé pour %s", req.InstanceID)
	return nil
}

func (s *Store) GetDashboardStats() (api.DashboardStats, error) {
	logger.DebugCtx("STORE", "GetDashboardStats appelé")

	var stats api.DashboardStats
	stats.GlobalMetrics = make(map[string]int64)

	queryCounts := `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE last_seen_at > NOW() - INTERVAL '30 days')
		FROM instances
	`
	if err := s.db.QueryRow(queryCounts).Scan(&stats.TotalInstances, &stats.ActiveInstances); err != nil {
		logger.ErrorCtx("STORE", "Erreur SQL GetDashboardStats (counts): %v", err)
		return stats, err
	}

	queryJson := `
		SELECT data
		FROM (
			SELECT DISTINCT ON (instance_id) data
			FROM snapshots
			ORDER BY instance_id, snapshot_at DESC
		) as latest
	`
	rows, err := s.db.Query(queryJson)
	if err != nil {
		logger.ErrorCtx("STORE", "Erreur SQL GetDashboardStats (snapshots): %v", err)
		return stats, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var rawJSON []byte
		if err := rows.Scan(&rawJSON); err != nil {
			logger.WarnCtx("STORE", "Erreur scan snapshot: %v", err)
			continue
		}

		var metrics map[string]interface{}
		if err := json.Unmarshal(rawJSON, &metrics); err != nil {
			logger.WarnCtx("STORE", "Erreur unmarshal metrics: %v", err)
			continue
		}

		for key, val := range metrics {
			switch v := val.(type) {
			case float64: // JSON décode les nombres en float64 par défaut
				stats.GlobalMetrics[key] += int64(v)
			case int:
				stats.GlobalMetrics[key] += int64(v)
				// On ignore les strings ou booléens pour l'agrégation
			}
		}
	}

	logger.InfoCtx("STORE", "Stats calculées: %d instances totales, %d actives", stats.TotalInstances, stats.ActiveInstances)
	return stats, nil
}

func (s *Store) ListInstances(limit int) ([]api.InstanceSummary, error) {
	logger.DebugCtx("STORE", "ListInstances appelé (limit=%d)", limit)

	query := `
		SELECT
			i.instance_id, i.app_name, i.app_version, i.environment, i.status, i.last_seen_at, i.deployment_mode,
			COALESCE(s.data, '{}'::jsonb)
		FROM instances i
		LEFT JOIN LATERAL (
			SELECT data FROM snapshots
			WHERE instance_id = i.instance_id
			ORDER BY snapshot_at DESC
			LIMIT 1
		) s ON true
		ORDER BY i.last_seen_at DESC
		LIMIT $1
	`
	rows, err := s.db.Query(query, limit)
	if err != nil {
		logger.ErrorCtx("STORE", "Erreur SQL ListInstances: %v", err)
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var list []api.InstanceSummary
	for rows.Next() {
		var i api.InstanceSummary
		var rawMetrics []byte

		if err := rows.Scan(&i.InstanceID, &i.AppName, &i.AppVersion, &i.Environment, &i.Status, &i.LastSeenAt, &i.DeploymentMode, &rawMetrics); err != nil {
			logger.ErrorCtx("STORE", "Erreur scan instance: %v", err)
			return nil, err
		}

		_ = json.Unmarshal(rawMetrics, &i.Metrics)
		list = append(list, i)
	}

	logger.InfoCtx("STORE", "Liste récupérée: %d instances", len(list))
	return list, nil
}

func (s *Store) GetMetricsTimeSeries(appName string, periodHours int) (map[string]interface{}, error) {
	logger.DebugCtx("STORE", "GetMetricsTimeSeries: app=%s, period=%dh", appName, periodHours)

	query := `
		SELECT s.snapshot_at, s.data
		FROM snapshots s
		JOIN instances i ON s.instance_id = i.instance_id
		WHERE i.app_name = $1
		  AND s.snapshot_at > NOW() - INTERVAL '1 hour' * $2
		ORDER BY s.snapshot_at ASC
	`

	rows, err := s.db.Query(query, appName, periodHours)
	if err != nil {
		logger.ErrorCtx("STORE", "Erreur SQL GetMetricsTimeSeries: %v", err)
		return nil, err
	}
	defer rows.Close()

	type dataPoint struct {
		timestamp string
		metrics   map[string]interface{}
	}

	var dataPoints []dataPoint
	for rows.Next() {
		var timestamp string
		var rawMetrics []byte

		if err := rows.Scan(&timestamp, &rawMetrics); err != nil {
			logger.WarnCtx("STORE", "Erreur scan snapshot: %v", err)
			continue
		}

		var metrics map[string]interface{}
		if err := json.Unmarshal(rawMetrics, &metrics); err != nil {
			logger.WarnCtx("STORE", "Erreur unmarshal metrics: %v", err)
			continue
		}

		dataPoints = append(dataPoints, dataPoint{
			timestamp: timestamp,
			metrics:   metrics,
		})
	}

	timestampMap := make(map[string]map[string]float64)
	var timestamps []string

	for _, dp := range dataPoints {
		if _, exists := timestampMap[dp.timestamp]; !exists {
			timestampMap[dp.timestamp] = make(map[string]float64)
			timestamps = append(timestamps, dp.timestamp)
		}

		for key, val := range dp.metrics {
			if v, ok := val.(float64); ok {
				timestampMap[dp.timestamp][key] += v
			}
		}
	}

	result := make(map[string]interface{})
	result["timestamps"] = timestamps

	metricsData := make(map[string][]float64)
	for _, ts := range timestamps {
		for metricKey, value := range timestampMap[ts] {
			metricsData[metricKey] = append(metricsData[metricKey], value)
		}
	}
	result["metrics"] = metricsData

	logger.InfoCtx("STORE", "TimeSeries récupérée: %d points pour %s", len(timestamps), appName)
	return result, nil
}
