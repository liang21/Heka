package e2e

// tasks.md: T156 | spec.md: §4.13 Health Check E2E Test
// Tests the health check endpoint that verifies connectivity to DB, Redis, Milvus

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/liang21/heka/internal/infrastructure/cache"
	"github.com/liang21/heka/internal/interface/http/handler"
)

// healthResponse is the expected JSON structure for health check responses
type healthResponse struct {
	Status string                 `json:"status"`
	Checks map[string]checkResult `json:"checks"`
}

type checkResult struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// setupHealthTest starts a testcontainers PostgreSQL and returns a test environment
func setupHealthTest(t *testing.T) (*sql.DB, *chi.Mux, func()) {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:15-alpine",
		tcpostgres.WithDatabase("heka_health_test"),
		tcpostgres.WithUsername("heka"),
		tcpostgres.WithPassword("heka_test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "AC: PostgreSQL testcontainers must start for health test")

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	gormDB, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := gormDB.DB()
	require.NoError(t, err)

	// Create health handler with DB (no Redis for simplicity)
	healthHandler := handler.NewHealthHandler(sqlDB, nil)

	r := chi.NewRouter()
	r.Get("/api/v1/health", healthHandler.Health)

	cleanup := func() {
		sqlDB.Close()
		pgContainer.Terminate(ctx)
	}

	return sqlDB, r, cleanup
}

// TestHealthCheck verifies the health check endpoint returns correct status
// spec.md §4.13 AC: GET /api/v1/health verifies connectivity to DB, Redis, Milvus
func TestHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping health test in short mode")
	}

	db, router, cleanup := setupHealthTest(t)
	defer cleanup()

	// Verify DB is actually connected
	require.NoError(t, db.Ping())

	server := httptest.NewServer(router)
	defer server.Close()

	// AC: GET /api/v1/health returns 200 with status information
	resp, err := http.Get(server.URL + "/api/v1/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode,
		"AC: Health check must return 200 when DB is healthy")

	var health healthResponse
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(t, err, "AC: Health response must be valid JSON")

	// AC: Top-level status must be "ok"
	assert.Equal(t, "ok", health.Status,
		"AC: Health check overall status must be 'ok' when dependencies are healthy")

	// AC: DB check must be present and healthy
	assert.Contains(t, health.Checks, "db",
		"AC: Health check must include 'db' check")
	assert.Equal(t, "ok", health.Checks["db"].Status,
		"AC: DB health check must report 'ok'")
}

// TestHealthCheckDBFailure verifies health check reports when DB is unreachable
func TestHealthCheckDBFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping health test in short mode")
	}

	// Create a health handler with a closed DB connection
	// First create a valid connection, then close it
	ctx := context.Background()
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:15-alpine",
		tcpostgres.WithDatabase("heka_health_fail_test"),
		tcpostgres.WithUsername("heka"),
		tcpostgres.WithPassword("heka_test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	dsn, _ := pgContainer.ConnectionString(ctx, "sslmode=disable")
	gormDB, _ := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	sqlDB, _ := gormDB.DB()

	// Close the DB connection to simulate failure
	sqlDB.Close()

	healthHandler := handler.NewHealthHandler(sqlDB, nil)

	r := chi.NewRouter()
	r.Get("/api/v1/health", healthHandler.Health)

	server := httptest.NewServer(r)
	defer server.Close()
	defer pgContainer.Terminate(ctx)

	resp, err := http.Get(server.URL + "/api/v1/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	// AC: Should return 503 Service Unavailable when DB is down
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode,
		"AC: Health check should return 503 when DB is down")

	var health healthResponse
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(t, err)

	assert.Equal(t, "degraded", health.Status,
		"AC: Health status should be 'degraded' when DB is down")
}
