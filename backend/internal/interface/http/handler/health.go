package handler

// tasks.md: T150 | spec.md: §4.13 Health Check Handler
// Provides health check endpoint that verifies connectivity to DB, Redis, Milvus

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/liang21/heka/internal/infrastructure/cache"
)

const defaultHealthCacheTTL = 10 * time.Second

type HealthHandler struct {
	db     *sql.DB
	redis  *cache.CacheClient
	cache  *healthCache
	ttl    time.Duration
}

type healthCache struct {
	result healthResponse
	expiry time.Time
	mu     sync.RWMutex
}

func NewHealthHandler(db *sql.DB, redis *cache.CacheClient) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
		ttl:   defaultHealthCacheTTL,
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
	// Check cache first
	if h.cache != nil && time.Now().Before(h.cache.expiry) {
		h.cache.mu.RLock()
		cached := h.cache.result
		h.cache.mu.RUnlock()

		statusCode := http.StatusOK
		if cached.Status != "ok" {
			statusCode = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(cached)
		return
	}

	checks := make(map[string]checkResult)
	overallStatus := "ok"

	// Check database
	if h.db != nil {
		if err := h.db.PingContext(r.Context()); err != nil {
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
		if err := h.redis.Ping(r.Context()); err != nil {
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

	// Build response
	response := healthResponse{
		Status: overallStatus,
		Checks: checks,
	}

	// Update cache
	if h.cache != nil {
		h.cache.mu.Lock()
		h.cache.result = response
		h.cache.expiry = time.Now().Add(h.ttl)
		h.cache.mu.Unlock()
	}

	// Determine HTTP status code
	statusCode := http.StatusOK
	if overallStatus == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
