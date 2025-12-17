// SPDX-License-Identifier: AGPL-3.0-or-later

package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/btouchard/shm/internal/config"
	"github.com/btouchard/shm/internal/middleware"
	"github.com/btouchard/shm/pkg/crypto"
)

// =============================================================================
// MOCK STORE
// =============================================================================

type mockStore struct {
	instances       map[string]mockInstance
	snapshots       []SnapshotRequest
	registerErr     error
	activateErr     error
	snapshotErr     error
	getKeyErr       error
	statsErr        error
	listErr         error
	metricsErr      error
	dashboardStats  DashboardStats
	instanceList    []InstanceSummary
	metricsData     map[string]interface{}
}

type mockInstance struct {
	publicKey string
	status    string
}

func newMockStore() *mockStore {
	return &mockStore{
		instances: make(map[string]mockInstance),
	}
}

func (m *mockStore) RegisterInstance(req RegisterRequest) error {
	if m.registerErr != nil {
		return m.registerErr
	}
	m.instances[req.InstanceID] = mockInstance{
		publicKey: req.PublicKey,
		status:    "pending",
	}
	return nil
}

func (m *mockStore) ActivateInstance(instanceID string) error {
	if m.activateErr != nil {
		return m.activateErr
	}
	if inst, ok := m.instances[instanceID]; ok {
		inst.status = "active"
		m.instances[instanceID] = inst
		return nil
	}
	return errors.New("instance not found")
}

func (m *mockStore) SaveSnapshot(req SnapshotRequest) error {
	if m.snapshotErr != nil {
		return m.snapshotErr
	}
	m.snapshots = append(m.snapshots, req)
	return nil
}

func (m *mockStore) GetInstanceKey(instanceID string) (string, error) {
	if m.getKeyErr != nil {
		return "", m.getKeyErr
	}
	if inst, ok := m.instances[instanceID]; ok {
		if inst.status == "revoked" {
			return "", errors.New("instance revoked")
		}
		return inst.publicKey, nil
	}
	return "", errors.New("instance not found")
}

func (m *mockStore) GetDashboardStats() (DashboardStats, error) {
	if m.statsErr != nil {
		return DashboardStats{}, m.statsErr
	}
	return m.dashboardStats, nil
}

func (m *mockStore) ListInstances(limit int) ([]InstanceSummary, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if limit > len(m.instanceList) {
		return m.instanceList, nil
	}
	return m.instanceList[:limit], nil
}

func (m *mockStore) GetMetricsTimeSeries(appName string, periodHours int) (map[string]interface{}, error) {
	if m.metricsErr != nil {
		return nil, m.metricsErr
	}
	return m.metricsData, nil
}

// =============================================================================
// SIGNED REQUEST MIDDLEWARE TESTS
// =============================================================================

func TestSignedRequestMiddleware_ValidSignature(t *testing.T) {
	store := newMockStore()

	// Generate keypair and register instance
	pub, priv, _ := crypto.GenerateKeypair()
	pubHex := hex.EncodeToString(pub)
	instanceID := "test-instance-123"

	store.instances[instanceID] = mockInstance{
		publicKey: pubHex,
		status:    "active",
	}

	// Create handler that tracks if it was called
	handlerCalled := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}

	// Create signed request
	body := []byte(`{"action": "test"}`)
	signature := crypto.Sign(priv, body)

	req := httptest.NewRequest("POST", "/v1/activate", bytes.NewBuffer(body))
	req.Header.Set("X-Instance-ID", instanceID)
	req.Header.Set("X-Signature", signature)

	rec := httptest.NewRecorder()

	// Apply middleware
	middleware := signedRequestMiddleware(store, handler)
	middleware(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !handlerCalled {
		t.Error("handler should have been called for valid signature")
	}
}

func TestSignedRequestMiddleware_MissingInstanceID(t *testing.T) {
	store := newMockStore()

	handler := func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}

	req := httptest.NewRequest("POST", "/v1/activate", bytes.NewBufferString(`{}`))
	req.Header.Set("X-Signature", "some-signature")
	// Missing X-Instance-ID

	rec := httptest.NewRecorder()

	middleware := signedRequestMiddleware(store, handler)
	middleware(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestSignedRequestMiddleware_MissingSignature(t *testing.T) {
	store := newMockStore()

	handler := func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}

	req := httptest.NewRequest("POST", "/v1/activate", bytes.NewBufferString(`{}`))
	req.Header.Set("X-Instance-ID", "test-instance")
	// Missing X-Signature

	rec := httptest.NewRecorder()

	middleware := signedRequestMiddleware(store, handler)
	middleware(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestSignedRequestMiddleware_InvalidSignature(t *testing.T) {
	store := newMockStore()

	pub, _, _ := crypto.GenerateKeypair()
	pubHex := hex.EncodeToString(pub)
	instanceID := "test-instance"

	store.instances[instanceID] = mockInstance{
		publicKey: pubHex,
		status:    "active",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for invalid signature")
	}

	req := httptest.NewRequest("POST", "/v1/activate", bytes.NewBufferString(`{}`))
	req.Header.Set("X-Instance-ID", instanceID)
	req.Header.Set("X-Signature", "invalid-signature-not-hex")

	rec := httptest.NewRecorder()

	middleware := signedRequestMiddleware(store, handler)
	middleware(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestSignedRequestMiddleware_WrongKey(t *testing.T) {
	store := newMockStore()

	// Register with one key
	pub1, _, _ := crypto.GenerateKeypair()
	// Sign with different key
	_, priv2, _ := crypto.GenerateKeypair()

	instanceID := "test-instance"
	store.instances[instanceID] = mockInstance{
		publicKey: hex.EncodeToString(pub1),
		status:    "active",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for wrong key")
	}

	body := []byte(`{"action": "test"}`)
	signature := crypto.Sign(priv2, body) // Signed with wrong key!

	req := httptest.NewRequest("POST", "/v1/activate", bytes.NewBuffer(body))
	req.Header.Set("X-Instance-ID", instanceID)
	req.Header.Set("X-Signature", signature)

	rec := httptest.NewRecorder()

	middleware := signedRequestMiddleware(store, handler)
	middleware(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestSignedRequestMiddleware_UnknownInstance(t *testing.T) {
	store := newMockStore()
	// Don't register any instance

	_, priv, _ := crypto.GenerateKeypair()

	handler := func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for unknown instance")
	}

	body := []byte(`{"action": "test"}`)
	signature := crypto.Sign(priv, body)

	req := httptest.NewRequest("POST", "/v1/activate", bytes.NewBuffer(body))
	req.Header.Set("X-Instance-ID", "unknown-instance")
	req.Header.Set("X-Signature", signature)

	rec := httptest.NewRecorder()

	middleware := signedRequestMiddleware(store, handler)
	middleware(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestSignedRequestMiddleware_RevokedInstance(t *testing.T) {
	store := newMockStore()

	pub, priv, _ := crypto.GenerateKeypair()
	instanceID := "revoked-instance"

	store.instances[instanceID] = mockInstance{
		publicKey: hex.EncodeToString(pub),
		status:    "revoked", // Instance is revoked!
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for revoked instance")
	}

	body := []byte(`{"action": "test"}`)
	signature := crypto.Sign(priv, body)

	req := httptest.NewRequest("POST", "/v1/activate", bytes.NewBuffer(body))
	req.Header.Set("X-Instance-ID", instanceID)
	req.Header.Set("X-Signature", signature)

	rec := httptest.NewRecorder()

	middleware := signedRequestMiddleware(store, handler)
	middleware(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestSignedRequestMiddleware_BodyPreservedAfterVerification(t *testing.T) {
	store := newMockStore()

	pub, priv, _ := crypto.GenerateKeypair()
	instanceID := "test-instance"

	store.instances[instanceID] = mockInstance{
		publicKey: hex.EncodeToString(pub),
		status:    "active",
	}

	originalBody := `{"key": "value", "number": 42}`
	var receivedBody string

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Try to read body in handler - it should still be available
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		receivedBody = buf.String()
		w.WriteHeader(http.StatusOK)
	}

	body := []byte(originalBody)
	signature := crypto.Sign(priv, body)

	req := httptest.NewRequest("POST", "/v1/snapshot", bytes.NewBuffer(body))
	req.Header.Set("X-Instance-ID", instanceID)
	req.Header.Set("X-Signature", signature)

	rec := httptest.NewRecorder()

	middleware := signedRequestMiddleware(store, handler)
	middleware(rec, req)

	if receivedBody != originalBody {
		t.Errorf("body not preserved: got %q, want %q", receivedBody, originalBody)
	}
}

func TestSignedRequestMiddleware_TamperedBody(t *testing.T) {
	store := newMockStore()

	pub, priv, _ := crypto.GenerateKeypair()
	instanceID := "test-instance"

	store.instances[instanceID] = mockInstance{
		publicKey: hex.EncodeToString(pub),
		status:    "active",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for tampered body")
	}

	// Sign original body
	originalBody := []byte(`{"amount": 100}`)
	signature := crypto.Sign(priv, originalBody)

	// But send tampered body
	tamperedBody := []byte(`{"amount": 1000000}`)

	req := httptest.NewRequest("POST", "/v1/snapshot", bytes.NewBuffer(tamperedBody))
	req.Header.Set("X-Instance-ID", instanceID)
	req.Header.Set("X-Signature", signature)

	rec := httptest.NewRecorder()

	middleware := signedRequestMiddleware(store, handler)
	middleware(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for tampered body, got %d", rec.Code)
	}
}

// =============================================================================
// ROUTER INTEGRATION TESTS
// =============================================================================

func TestNewRouter_RoutesRegistered(t *testing.T) {
	store := newMockStore()
	rl := middleware.NewRateLimiter(config.RateLimitConfig{Enabled: false})
	defer rl.Stop()

	router := NewRouter(store, rl)

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/healthcheck"},
		{"POST", "/v1/register"},
		// Note: /v1/activate and /v1/snapshot require auth, tested separately
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			// Should not be 404
			if rec.Code == http.StatusNotFound {
				t.Errorf("route %s %s returned 404", route.method, route.path)
			}
		})
	}
}

func TestRouter_HealthcheckEndpoint(t *testing.T) {
	store := newMockStore()
	rl := middleware.NewRateLimiter(config.RateLimitConfig{Enabled: false})
	defer rl.Stop()

	router := NewRouter(store, rl)

	req := httptest.NewRequest("GET", "/api/v1/healthcheck", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("healthcheck returned %d, want 200", rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("healthcheck status = %q, want 'ok'", resp["status"])
	}
}

func TestRouter_RegisterEndpoint(t *testing.T) {
	store := newMockStore()
	rl := middleware.NewRateLimiter(config.RateLimitConfig{Enabled: false})
	defer rl.Stop()

	router := NewRouter(store, rl)

	reqBody := RegisterRequest{
		InstanceID:  "new-instance",
		PublicKey:   "abcd1234",
		AppName:     "test-app",
		AppVersion:  "1.0.0",
		Environment: "test",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("register returned %d, want 201", rec.Code)
	}

	// Verify instance was stored
	if _, ok := store.instances["new-instance"]; !ok {
		t.Error("instance was not stored")
	}
}

func TestRouter_RegisterEndpoint_InvalidJSON(t *testing.T) {
	store := newMockStore()
	rl := middleware.NewRateLimiter(config.RateLimitConfig{Enabled: false})
	defer rl.Stop()

	router := NewRouter(store, rl)

	req := httptest.NewRequest("POST", "/v1/register", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("invalid JSON should return 400, got %d", rec.Code)
	}
}

func TestRouter_RegisterEndpoint_WrongMethod(t *testing.T) {
	store := newMockStore()
	rl := middleware.NewRateLimiter(config.RateLimitConfig{Enabled: false})
	defer rl.Stop()

	router := NewRouter(store, rl)

	req := httptest.NewRequest("GET", "/v1/register", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET on POST endpoint should return 405, got %d", rec.Code)
	}
}

func TestRouter_FullFlow_RegisterActivateSnapshot(t *testing.T) {
	store := newMockStore()
	rl := middleware.NewRateLimiter(config.RateLimitConfig{Enabled: false})
	defer rl.Stop()

	router := NewRouter(store, rl)

	// 1. Generate keypair
	pub, priv, _ := crypto.GenerateKeypair()
	pubHex := hex.EncodeToString(pub)
	instanceID := "integration-test-instance"

	// 2. Register
	regReq := RegisterRequest{
		InstanceID: instanceID,
		PublicKey:  pubHex,
		AppName:    "test-app",
		AppVersion: "1.0.0",
	}
	regBody, _ := json.Marshal(regReq)

	req := httptest.NewRequest("POST", "/v1/register", bytes.NewBuffer(regBody))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("register failed: %d", rec.Code)
	}

	// 3. Activate with signed request
	activateBody := []byte(`{"action": "activate"}`)
	activateSig := crypto.Sign(priv, activateBody)

	req = httptest.NewRequest("POST", "/v1/activate", bytes.NewBuffer(activateBody))
	req.Header.Set("X-Instance-ID", instanceID)
	req.Header.Set("X-Signature", activateSig)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("activate failed: %d", rec.Code)
	}

	// 4. Send snapshot with signed request
	snapshotReq := SnapshotRequest{
		InstanceID: instanceID,
		Timestamp:  time.Now(),
		Metrics:    json.RawMessage(`{"cpu": 50, "mem": 1024}`),
	}
	snapshotBody, _ := json.Marshal(snapshotReq)
	snapshotSig := crypto.Sign(priv, snapshotBody)

	req = httptest.NewRequest("POST", "/v1/snapshot", bytes.NewBuffer(snapshotBody))
	req.Header.Set("X-Instance-ID", instanceID)
	req.Header.Set("X-Signature", snapshotSig)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("snapshot failed: %d", rec.Code)
	}

	// Verify snapshot was stored
	if len(store.snapshots) != 1 {
		t.Errorf("expected 1 snapshot stored, got %d", len(store.snapshots))
	}
}

// =============================================================================
// ADMIN ENDPOINTS TESTS
// =============================================================================

func TestRouter_AdminStats(t *testing.T) {
	store := newMockStore()
	store.dashboardStats = DashboardStats{
		TotalInstances:  100,
		ActiveInstances: 42,
		GlobalMetrics:   map[string]int64{"requests": 1000},
	}

	rl := middleware.NewRateLimiter(config.RateLimitConfig{Enabled: false})
	defer rl.Stop()

	router := NewRouter(store, rl)

	req := httptest.NewRequest("GET", "/api/v1/admin/stats", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("admin stats returned %d, want 200", rec.Code)
	}

	var stats DashboardStats
	json.NewDecoder(rec.Body).Decode(&stats)

	if stats.TotalInstances != 100 {
		t.Errorf("TotalInstances = %d, want 100", stats.TotalInstances)
	}
}

func TestRouter_AdminInstances(t *testing.T) {
	store := newMockStore()
	store.instanceList = []InstanceSummary{
		{InstanceID: "inst-1", AppName: "app1"},
		{InstanceID: "inst-2", AppName: "app2"},
	}

	rl := middleware.NewRateLimiter(config.RateLimitConfig{Enabled: false})
	defer rl.Stop()

	router := NewRouter(store, rl)

	req := httptest.NewRequest("GET", "/api/v1/admin/instances", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("admin instances returned %d, want 200", rec.Code)
	}

	var list []InstanceSummary
	json.NewDecoder(rec.Body).Decode(&list)

	if len(list) != 2 {
		t.Errorf("got %d instances, want 2", len(list))
	}
}

func TestRouter_AdminMetrics_PeriodParsing(t *testing.T) {
	store := newMockStore()
	store.metricsData = map[string]interface{}{
		"timestamps": []string{"2024-01-01"},
		"metrics":    map[string][]float64{"cpu": {50.0}},
	}

	rl := middleware.NewRateLimiter(config.RateLimitConfig{Enabled: false})
	defer rl.Stop()

	router := NewRouter(store, rl)

	tests := []struct {
		period string
	}{
		{""},      // default 24h
		{"7d"},    // 7 days
		{"30d"},   // 30 days
		{"3m"},    // 3 months
		{"1y"},    // 1 year
		{"all"},   // all time
		{"bogus"}, // invalid falls back to default
	}

	for _, tt := range tests {
		t.Run("period="+tt.period, func(t *testing.T) {
			url := "/api/v1/admin/metrics/testapp"
			if tt.period != "" {
				url += "?period=" + tt.period
			}

			req := httptest.NewRequest("GET", url, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("period=%s returned %d, want 200", tt.period, rec.Code)
			}
		})
	}
}

func TestRouter_AdminMetrics_MissingAppName(t *testing.T) {
	store := newMockStore()
	rl := middleware.NewRateLimiter(config.RateLimitConfig{Enabled: false})
	defer rl.Stop()

	router := NewRouter(store, rl)

	req := httptest.NewRequest("GET", "/api/v1/admin/metrics/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("missing app name should return 400, got %d", rec.Code)
	}
}
