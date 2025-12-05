package handlers

import (
	"context"
	"time"

	"github.com/Sambit-Mondal/karbos/server/internal/database"
	"github.com/Sambit-Mondal/karbos/server/internal/queue"
	"github.com/gofiber/fiber/v2"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db    *database.DB
	queue *queue.RedisQueue
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *database.DB, queue *queue.RedisQueue) *HealthHandler {
	return &HealthHandler{
		db:    db,
		queue: queue,
	}
}

// HealthCheck handles GET /health
func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Check database
	dbHealthy := true
	if err := h.db.HealthCheck(); err != nil {
		dbHealthy = false
	}

	// Check Redis
	redisHealthy := true
	if err := h.queue.HealthCheck(ctx); err != nil {
		redisHealthy = false
	}

	// Get queue stats
	immediateQueueLength, _ := h.queue.GetImmediateQueueLength(ctx)
	delayedQueueLength, _ := h.queue.GetDelayedQueueLength(ctx)

	status := fiber.StatusOK
	if !dbHealthy || !redisHealthy {
		status = fiber.StatusServiceUnavailable
	}

	return c.Status(status).JSON(fiber.Map{
		"status": map[string]bool{
			"database": dbHealthy,
			"redis":    redisHealthy,
			"healthy":  dbHealthy && redisHealthy,
		},
		"queue": fiber.Map{
			"immediate": immediateQueueLength,
			"delayed":   delayedQueueLength,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// ReadyCheck handles GET /ready
func (h *HealthHandler) ReadyCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Quick health checks
	if err := h.db.HealthCheck(); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"ready":   false,
			"message": "Database not ready",
		})
	}

	if err := h.queue.HealthCheck(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"ready":   false,
			"message": "Redis not ready",
		})
	}

	return c.JSON(fiber.Map{
		"ready":     true,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
