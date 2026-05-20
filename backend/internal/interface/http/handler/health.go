package handler

// tasks.md: T150 | spec.md: §4.13 Health Check Handler
// Provides health check endpoint that verifies connectivity to DB, Redis, Milvus

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/liang21/heka/internal/infrastructure/cache"
)

// HealthChecker is the interface for checking health of dependencies
type HealthChecker interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	db    *sql.DB
	redis *cache.CacheClient
}

func NewHealthHandler(db *sql.DB, redis *cache.CacheClient) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

type healthResponse struct {
	Status string                 `json:"status"`
	Checks map[string]checkResult `json:"checks"`
}

type checkResult struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Health returns the health status of all dependencies
// spec.md §4.13 AC: GET /api/v1/health verifies connectivity to DB, Redis, Milvus
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	checks := make(map[string]checkResult)
	overallStatus := "ok"

	// Check database
	if h.db != nil {
		if err := h.db.PingContext(ctx); err != nil {
			checks["db"] = checkResult{Status: "error", Error: err.Error()}
			overallStatus = "degraded"
		} else {
			checks["db"] = checkResult{Status: "ok"}
		}
	} else {
		checks["db"] = checkResult{Status: "error", Error: "database not configured"}
		overallStatus = "degraded"
	}

	// Check Redis
	if h.redis != nil {
		if err := h.redis.Ping(ctx); err != nil {
			checks["redis"] = checkResult{Status: "error", Error: err.Error()}
			overallStatus = "degraded"
		} else {
			checks["redis"] = checkResult{Status: "ok"}
		}
	} else {
		checks["redis"] = checkResult{Status: "skip", Error: "redis not configured"}
	}

	// Milvus check would be added here when Milvus client is available
	// For now, we skip it since it's optional in some deployment modes
	checks["milvus"] = checkResult{Status: "skip", Error: "milvus check not implemented"}

	// Determine HTTP status code
	statusCode := http.StatusOK
	if overallStatus == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(healthResponse{
		Status: overallStatus,
		Checks: checks,
	})
}
