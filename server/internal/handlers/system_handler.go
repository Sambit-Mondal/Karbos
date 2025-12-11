package handlers

import (
	"context"
	"time"

	"github.com/Sambit-Mondal/karbos/server/internal/models"
	"github.com/Sambit-Mondal/karbos/server/internal/queue"
	"github.com/gofiber/fiber/v2"
)

type SystemHandler struct {
	queue *queue.RedisQueue
}

func NewSystemHandler(queue *queue.RedisQueue) *SystemHandler {
	return &SystemHandler{
		queue: queue,
	}
}

// GetSystemHealth returns the system health status
// GET /api/system/health
func (h *SystemHandler) GetSystemHealth(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get active workers
	workers, err := h.queue.GetActiveWorkers(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get active workers",
		})
	}

	// Get queue depths
	immediateDepth, err := h.queue.GetImmediateQueueLength(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get immediate queue depth",
		})
	}

	delayedDepth, err := h.queue.GetDelayedQueueLength(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get delayed queue depth",
		})
	}

	// Calculate mock Redis latency (ping)
	start := time.Now()
	if err := h.queue.HealthCheck(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Redis health check failed",
		})
	}
	latencyMs := time.Since(start).Milliseconds()

	response := models.SystemHealthResponse{
		ActiveWorkers:       len(workers),
		WorkerIDs:           workers,
		QueueDepthImmediate: int(immediateDepth),
		QueueDepthDelayed:   int(delayedDepth),
		RedisLatencyMs:      int(latencyMs),
		Timestamp:           time.Now(),
	}

	return c.JSON(response)
}
