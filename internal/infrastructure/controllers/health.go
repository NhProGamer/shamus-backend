package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string            `json:"status"` // "healthy" or "unhealthy"
	Checks map[string]string `json:"checks"` // Individual component checks
}

// HealthHandler handles /health requests
func HealthHandler(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := HealthResponse{
			Status: "healthy",
			Checks: make(map[string]string),
		}

		// Check Redis connection
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := rdb.Ping(ctx).Err(); err != nil {
			response.Status = "unhealthy"
			response.Checks["redis"] = "unhealthy: " + err.Error()
			c.JSON(http.StatusServiceUnavailable, response)
			return
		}

		response.Checks["redis"] = "healthy"

		c.JSON(http.StatusOK, response)
	}
}

// ReadinessHandler handles /ready requests (for Kubernetes readiness probes)
func ReadinessHandler(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if Redis is ready
		ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
		defer cancel()

		if err := rdb.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"reason": "redis unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	}
}

// LivenessHandler handles /live requests (for Kubernetes liveness probes)
func LivenessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simple liveness check - server is alive if it can respond
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
		})
	}
}
