// SPDX-License-Identifier: AGPL-3.0-or-later

package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/btouchard/shm/pkg/api"
)

// =============================================================================
// TEST HELPERS
// =============================================================================

func newMockStore(t *testing.T) (*Store, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return &Store{db: db}, mock
}

// =============================================================================
// REGISTER INSTANCE TESTS
// =============================================================================

func TestRegisterInstance_NewInstance(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	req := api.RegisterRequest{
		InstanceID:     "test-instance-123",
		PublicKey:      "abc123pubkey",
		AppName:        "test-app",
		AppVersion:     "1.0.0",
		DeploymentMode: "docker",
		Environment:    "production",
		OSArch:         "linux/amd64",
	}

	mock.ExpectExec("INSERT INTO instances").
		WithArgs(req.InstanceID, req.PublicKey, req.AppName, req.AppVersion, req.DeploymentMode, req.Environment, req.OSArch).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.RegisterInstance(req)
	if err != nil {
		t.Errorf("RegisterInstance() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRegisterInstance_UpdateExisting(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	req := api.RegisterRequest{
		InstanceID: "existing-instance",
		PublicKey:  "newpubkey",
		AppName:    "updated-app",
		AppVersion: "2.0.0",
	}

	// ON CONFLICT DO UPDATE should also return success
	mock.ExpectExec("INSERT INTO instances").
		WithArgs(req.InstanceID, req.PublicKey, req.AppName, req.AppVersion, req.DeploymentMode, req.Environment, req.OSArch).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 0 insert, 1 affected (update)

	err := store.RegisterInstance(req)
	if err != nil {
		t.Errorf("RegisterInstance() update error = %v", err)
	}
}

func TestRegisterInstance_DBError(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	req := api.RegisterRequest{
		InstanceID: "test-instance",
		PublicKey:  "key",
	}

	mock.ExpectExec("INSERT INTO instances").
		WillReturnError(errors.New("connection refused"))

	err := store.RegisterInstance(req)
	if err == nil {
		t.Error("RegisterInstance() should return error on DB failure")
	}
}

// =============================================================================
// ACTIVATE INSTANCE TESTS
// =============================================================================

func TestActivateInstance_Success(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	instanceID := "test-instance"

	mock.ExpectExec("UPDATE instances SET status").
		WithArgs(instanceID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := store.ActivateInstance(instanceID)
	if err != nil {
		t.Errorf("ActivateInstance() error = %v", err)
	}
}

func TestActivateInstance_NotFound(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	instanceID := "non-existent"

	mock.ExpectExec("UPDATE instances SET status").
		WithArgs(instanceID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	err := store.ActivateInstance(instanceID)
	if err == nil {
		t.Error("ActivateInstance() should return error when instance not found")
	}
	if err.Error() != "instance not found" {
		t.Errorf("error = %q, want 'instance not found'", err.Error())
	}
}

func TestActivateInstance_DBError(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	mock.ExpectExec("UPDATE instances").
		WillReturnError(errors.New("db error"))

	err := store.ActivateInstance("test")
	if err == nil {
		t.Error("ActivateInstance() should return error on DB failure")
	}
}

// =============================================================================
// GET INSTANCE KEY TESTS
// =============================================================================

func TestGetInstanceKey_Success(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	instanceID := "test-instance"
	expectedKey := "public-key-hex-123"

	rows := sqlmock.NewRows([]string{"public_key"}).
		AddRow(expectedKey)

	mock.ExpectQuery("SELECT public_key FROM instances").
		WithArgs(instanceID).
		WillReturnRows(rows)

	key, err := store.GetInstanceKey(instanceID)
	if err != nil {
		t.Errorf("GetInstanceKey() error = %v", err)
	}
	if key != expectedKey {
		t.Errorf("key = %q, want %q", key, expectedKey)
	}
}

func TestGetInstanceKey_NotFound(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	mock.ExpectQuery("SELECT public_key FROM instances").
		WithArgs("unknown").
		WillReturnError(sql.ErrNoRows)

	key, err := store.GetInstanceKey("unknown")
	if err == nil {
		t.Error("GetInstanceKey() should return error for unknown instance")
	}
	if key != "" {
		t.Errorf("key should be empty for unknown instance, got %q", key)
	}
}

func TestGetInstanceKey_RevokedInstance(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	// Query excludes revoked instances, so it returns no rows
	mock.ExpectQuery("SELECT public_key FROM instances").
		WithArgs("revoked-instance").
		WillReturnError(sql.ErrNoRows)

	key, err := store.GetInstanceKey("revoked-instance")
	if err == nil {
		t.Error("GetInstanceKey() should return error for revoked instance")
	}
	if key != "" {
		t.Errorf("key should be empty for revoked instance")
	}
}

// =============================================================================
// SAVE SNAPSHOT TESTS
// =============================================================================

func TestSaveSnapshot_Success(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	req := api.SnapshotRequest{
		InstanceID: "test-instance",
		Timestamp:  time.Now(),
		Metrics:    json.RawMessage(`{"cpu": 50, "memory": 1024}`),
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO snapshots").
		WithArgs(req.InstanceID, req.Timestamp, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE instances SET last_seen_at").
		WithArgs(req.InstanceID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := store.SaveSnapshot(req)
	if err != nil {
		t.Errorf("SaveSnapshot() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestSaveSnapshot_InsertError_Rollback(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	req := api.SnapshotRequest{
		InstanceID: "test-instance",
		Timestamp:  time.Now(),
		Metrics:    json.RawMessage(`{}`),
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO snapshots").
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	err := store.SaveSnapshot(req)
	if err == nil {
		t.Error("SaveSnapshot() should return error on insert failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("rollback not called: %v", err)
	}
}

func TestSaveSnapshot_UpdateError_Rollback(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	req := api.SnapshotRequest{
		InstanceID: "test-instance",
		Timestamp:  time.Now(),
		Metrics:    json.RawMessage(`{}`),
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO snapshots").
		WithArgs(req.InstanceID, req.Timestamp, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE instances").
		WillReturnError(errors.New("update failed"))
	mock.ExpectRollback()

	err := store.SaveSnapshot(req)
	if err == nil {
		t.Error("SaveSnapshot() should return error on update failure")
	}
}

func TestSaveSnapshot_CommitError(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	req := api.SnapshotRequest{
		InstanceID: "test-instance",
		Timestamp:  time.Now(),
		Metrics:    json.RawMessage(`{}`),
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO snapshots").
		WithArgs(req.InstanceID, req.Timestamp, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE instances").
		WithArgs(req.InstanceID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	err := store.SaveSnapshot(req)
	if err == nil {
		t.Error("SaveSnapshot() should return error on commit failure")
	}
}

// =============================================================================
// GET DASHBOARD STATS TESTS
// =============================================================================

func TestGetDashboardStats_Success(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	// First query: counts
	countRows := sqlmock.NewRows([]string{"total", "active"}).
		AddRow(100, 42)
	mock.ExpectQuery("SELECT.*COUNT").
		WillReturnRows(countRows)

	// Second query: latest snapshots for aggregation
	snapshotRows := sqlmock.NewRows([]string{"data"}).
		AddRow(`{"requests": 100, "errors": 5}`).
		AddRow(`{"requests": 200, "errors": 10}`)
	mock.ExpectQuery("SELECT data").
		WillReturnRows(snapshotRows)

	stats, err := store.GetDashboardStats()
	if err != nil {
		t.Errorf("GetDashboardStats() error = %v", err)
	}

	if stats.TotalInstances != 100 {
		t.Errorf("TotalInstances = %d, want 100", stats.TotalInstances)
	}
	if stats.ActiveInstances != 42 {
		t.Errorf("ActiveInstances = %d, want 42", stats.ActiveInstances)
	}
	if stats.GlobalMetrics["requests"] != 300 {
		t.Errorf("requests = %d, want 300", stats.GlobalMetrics["requests"])
	}
	if stats.GlobalMetrics["errors"] != 15 {
		t.Errorf("errors = %d, want 15", stats.GlobalMetrics["errors"])
	}
}

func TestGetDashboardStats_EmptyDB(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	countRows := sqlmock.NewRows([]string{"total", "active"}).
		AddRow(0, 0)
	mock.ExpectQuery("SELECT.*COUNT").
		WillReturnRows(countRows)

	snapshotRows := sqlmock.NewRows([]string{"data"}) // No rows
	mock.ExpectQuery("SELECT data").
		WillReturnRows(snapshotRows)

	stats, err := store.GetDashboardStats()
	if err != nil {
		t.Errorf("GetDashboardStats() error = %v", err)
	}

	if stats.TotalInstances != 0 {
		t.Errorf("TotalInstances = %d, want 0", stats.TotalInstances)
	}
}

func TestGetDashboardStats_CountQueryError(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	mock.ExpectQuery("SELECT.*COUNT").
		WillReturnError(errors.New("db error"))

	_, err := store.GetDashboardStats()
	if err == nil {
		t.Error("GetDashboardStats() should return error on count query failure")
	}
}

func TestGetDashboardStats_InvalidJSON(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	countRows := sqlmock.NewRows([]string{"total", "active"}).
		AddRow(1, 1)
	mock.ExpectQuery("SELECT.*COUNT").
		WillReturnRows(countRows)

	// Invalid JSON should be skipped, not cause error
	snapshotRows := sqlmock.NewRows([]string{"data"}).
		AddRow(`not valid json`).
		AddRow(`{"valid": 100}`)
	mock.ExpectQuery("SELECT data").
		WillReturnRows(snapshotRows)

	stats, err := store.GetDashboardStats()
	if err != nil {
		t.Errorf("GetDashboardStats() should skip invalid JSON: %v", err)
	}

	// Should still have aggregated the valid row
	if stats.GlobalMetrics["valid"] != 100 {
		t.Errorf("valid metric not aggregated correctly")
	}
}

// =============================================================================
// LIST INSTANCES TESTS
// =============================================================================

func TestListInstances_Success(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	rows := sqlmock.NewRows([]string{
		"instance_id", "app_name", "app_version", "environment",
		"status", "last_seen_at", "deployment_mode", "data",
	}).
		AddRow("inst-1", "app1", "1.0", "prod", "active", time.Now(), "docker", `{"cpu": 50}`).
		AddRow("inst-2", "app2", "2.0", "staging", "pending", time.Now(), "k8s", `{}`)

	mock.ExpectQuery("SELECT.*FROM instances").
		WithArgs(10).
		WillReturnRows(rows)

	list, err := store.ListInstances(10)
	if err != nil {
		t.Errorf("ListInstances() error = %v", err)
	}

	if len(list) != 2 {
		t.Errorf("got %d instances, want 2", len(list))
	}
	if list[0].InstanceID != "inst-1" {
		t.Errorf("first instance ID = %q, want 'inst-1'", list[0].InstanceID)
	}
}

func TestListInstances_LimitRespected(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	// Only return 5 rows even though we have more
	rows := sqlmock.NewRows([]string{
		"instance_id", "app_name", "app_version", "environment",
		"status", "last_seen_at", "deployment_mode", "data",
	})
	for i := 0; i < 5; i++ {
		rows.AddRow("inst", "app", "1.0", "prod", "active", time.Now(), "docker", `{}`)
	}

	mock.ExpectQuery("SELECT.*FROM instances.*LIMIT").
		WithArgs(5).
		WillReturnRows(rows)

	list, err := store.ListInstances(5)
	if err != nil {
		t.Errorf("ListInstances() error = %v", err)
	}

	if len(list) != 5 {
		t.Errorf("got %d instances, limit was 5", len(list))
	}
}

func TestListInstances_DBError(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	mock.ExpectQuery("SELECT.*FROM instances").
		WillReturnError(errors.New("db error"))

	_, err := store.ListInstances(10)
	if err == nil {
		t.Error("ListInstances() should return error on DB failure")
	}
}

// =============================================================================
// GET METRICS TIME SERIES TESTS
// =============================================================================

func TestGetMetricsTimeSeries_Success(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	rows := sqlmock.NewRows([]string{"snapshot_at", "data"}).
		AddRow("2024-01-01T00:00:00Z", `{"cpu": 50}`).
		AddRow("2024-01-01T01:00:00Z", `{"cpu": 60}`).
		AddRow("2024-01-01T01:00:00Z", `{"cpu": 40}`) // Same timestamp, different instance

	mock.ExpectQuery("SELECT s.snapshot_at, s.data").
		WithArgs("test-app", 24).
		WillReturnRows(rows)

	result, err := store.GetMetricsTimeSeries("test-app", 24)
	if err != nil {
		t.Errorf("GetMetricsTimeSeries() error = %v", err)
	}

	timestamps, ok := result["timestamps"].([]string)
	if !ok {
		t.Fatal("timestamps should be []string")
	}
	if len(timestamps) != 2 {
		t.Errorf("got %d timestamps, want 2 (unique)", len(timestamps))
	}

	metrics, ok := result["metrics"].(map[string][]float64)
	if !ok {
		t.Fatal("metrics should be map[string][]float64")
	}
	cpuMetrics := metrics["cpu"]
	if len(cpuMetrics) != 2 {
		t.Errorf("got %d cpu data points, want 2", len(cpuMetrics))
	}
	// Second timestamp should aggregate: 60 + 40 = 100
	if cpuMetrics[1] != 100 {
		t.Errorf("aggregated cpu = %f, want 100", cpuMetrics[1])
	}
}

func TestGetMetricsTimeSeries_EmptyResult(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	rows := sqlmock.NewRows([]string{"snapshot_at", "data"})
	mock.ExpectQuery("SELECT s.snapshot_at, s.data").
		WithArgs("unknown-app", 24).
		WillReturnRows(rows)

	result, err := store.GetMetricsTimeSeries("unknown-app", 24)
	if err != nil {
		t.Errorf("GetMetricsTimeSeries() error = %v", err)
	}

	timestamps := result["timestamps"].([]string)
	if len(timestamps) != 0 {
		t.Errorf("got %d timestamps, want 0 for unknown app", len(timestamps))
	}
}

func TestGetMetricsTimeSeries_DBError(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	mock.ExpectQuery("SELECT").
		WillReturnError(errors.New("db error"))

	_, err := store.GetMetricsTimeSeries("app", 24)
	if err == nil {
		t.Error("GetMetricsTimeSeries() should return error on DB failure")
	}
}

// =============================================================================
// TRANSACTION ISOLATION TESTS
// =============================================================================

func TestSaveSnapshot_TransactionIsolation(t *testing.T) {
	store, mock := newMockStore(t)
	defer store.db.Close()

	req := api.SnapshotRequest{
		InstanceID: "test-instance",
		Timestamp:  time.Now(),
		Metrics:    json.RawMessage(`{}`),
	}

	// Verify transaction is started
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO snapshots").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE instances").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	store.SaveSnapshot(req)

	// All expectations should be met in order
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("transaction order not respected: %v", err)
	}
}
