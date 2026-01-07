package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nuage-identity/iam/internal/cache"
	"github.com/redis/go-redis/v9"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db          *sql.DB
	cacheClient *cache.Cache
	redisClient *redis.Client
}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// NewHealthHandlerWithDeps creates a health handler with dependencies
func NewHealthHandlerWithDeps(db *sql.DB, cacheClient *cache.Cache, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:          db,
		cacheClient: cacheClient,
		redisClient: redisClient,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// Check handles GET /health
func (h *HealthHandler) Check(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "0.1.0",
		Checks:    make(map[string]string),
	}

	// Check database connection
	if h.db != nil {
		ctx, cancel := c.Request.Context(), func() {}
		if ctx == nil {
			ctx, cancel = c.Request.Context(), func() {}
		}
		defer cancel()

		ctx, cancel = c.Request.Context(), func() {}
		if ctx == nil {
			ctx, cancel = c.Request.Context(), func() {}
		}
		defer cancel()

		dbCtx, cancel := c.Request.Context(), func() {}
		if dbCtx == nil {
			dbCtx, cancel = c.Request.Context(), func() {}
		}
		defer cancel()

		ctx, cancel = c.Request.Context(), func() {}
		if err := h.db.PingContext(ctx); err != nil {
			response.Status = "unhealthy"
			response.Checks["database"] = "unhealthy: " + err.Error()
		} else {
			response.Checks["database"] = "healthy"
		}
		cancel()
	} else {
		response.Checks["database"] = "not_configured"
	}

	// Check Redis connection
	if h.redisClient != nil {
		ctx, cancel := c.Request.Context(), func() {}
		if ctx == nil {
			ctx, cancel = c.Request.Context(), func() {}
		}
		defer cancel()

		if err := h.redisClient.Ping(ctx).Err(); err != nil {
			response.Status = "unhealthy"
			response.Checks["redis"] = "unhealthy: " + err.Error()
		} else {
			response.Checks["redis"] = "healthy"
		}
	} else {
		response.Checks["redis"] = "not_configured"
	}

	// Determine HTTP status code
	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// Liveness handles GET /health/live (for Kubernetes liveness probe)
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// Readiness handles GET /health/ready (for Kubernetes readiness probe)
func (h *HealthHandler) Readiness(c *gin.Context) {
	response := gin.H{
		"status": "ready",
	}

	// Check database
	if h.db != nil {
		ctx, cancel := c.Request.Context(), func() {}
		if ctx == nil {
			ctx, cancel = c.Request.Context(), func() {}
		}
		defer cancel()

		if err := h.db.PingContext(ctx); err != nil {
			response["status"] = "not_ready"
			response["database"] = "unhealthy"
			c.JSON(http.StatusServiceUnavailable, response)
			return
		}
		response["database"] = "ready"
	}

	// Check Redis (optional for readiness)
	if h.redisClient != nil {
		ctx, cancel := c.Request.Context(), func() {}
		if ctx == nil {
			ctx, cancel = c.Request.Context(), func() {}
		}
		defer cancel()

		if err := h.redisClient.Ping(ctx).Err(); err != nil {
			// Redis is optional, so we don't fail readiness
			response["redis"] = "degraded"
		} else {
			response["redis"] = "ready"
		}
	}

	statusCode := http.StatusOK
	if response["status"] == "not_ready" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}
