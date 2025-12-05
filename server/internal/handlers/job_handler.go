package handlers

import (
	"context"
	"log"
	"time"

	"github.com/Sambit-Mondal/karbos/server/internal/database"
	"github.com/Sambit-Mondal/karbos/server/internal/models"
	"github.com/Sambit-Mondal/karbos/server/internal/queue"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// JobHandler handles job-related HTTP requests
type JobHandler struct {
	jobRepo *database.JobRepository
	queue   *queue.RedisQueue
}

// NewJobHandler creates a new job handler
func NewJobHandler(jobRepo *database.JobRepository, queue *queue.RedisQueue) *JobHandler {
	return &JobHandler{
		jobRepo: jobRepo,
		queue:   queue,
	}
}

// SubmitJob handles POST /api/submit
func (h *JobHandler) SubmitJob(c *fiber.Ctx) error {
	var req models.SubmitJobRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    fiber.StatusBadRequest,
		})
	}

	// Validate required fields
	if req.UserID == "" || req.DockerImage == "" || req.Deadline == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "validation_error",
			Message: "user_id, docker_image, and deadline are required",
			Code:    fiber.StatusBadRequest,
		})
	}

	// Parse deadline
	deadline, err := time.Parse(time.RFC3339, req.Deadline)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalid_deadline",
			Message: "Deadline must be in ISO 8601 format (e.g., 2025-12-05T18:00:00Z)",
			Code:    fiber.StatusBadRequest,
		})
	}

	// Validate deadline is in the future
	if deadline.Before(time.Now()) {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalid_deadline",
			Message: "Deadline must be in the future",
			Code:    fiber.StatusBadRequest,
		})
	}

	// Create job object
	job := &models.Job{
		ID:                uuid.New(),
		UserID:            req.UserID,
		DockerImage:       req.DockerImage,
		Command:           req.Command,
		Status:            models.JobStatusPending,
		Deadline:          deadline,
		EstimatedDuration: req.EstimatedDuration,
		Region:            req.Region,
		CreatedAt:         time.Now(),
		Metadata:          "{}",
	}

	// Save to database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.jobRepo.CreateJob(ctx, job); err != nil {
		log.Printf("Failed to create job in database: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to create job",
			Code:    fiber.StatusInternalServerError,
		})
	}

	log.Printf("✓ Created job in database: %s", job.ID)

	// Create queue item
	queueItem := &queue.QueueItem{
		JobID:         job.ID.String(),
		DockerImage:   job.DockerImage,
		Command:       job.Command,
		ScheduledTime: time.Now(), // Immediate execution for now
		Priority:      0,
	}

	// Push to Redis immediate queue
	if err := h.queue.EnqueueImmediate(ctx, queueItem); err != nil {
		log.Printf("Failed to enqueue job: %v", err)
		// Note: Job is already in database, so we return success but log the error
		// In production, you might want to implement retry logic or dead letter queue
	}

	// Return success response
	response := models.SubmitJobResponse{
		JobID:     job.ID.String(),
		Status:    job.Status,
		CreatedAt: job.CreatedAt,
		Message:   "Job submitted successfully",
	}

	log.Printf("✓ Job submitted successfully: %s (UserID: %s, Image: %s)",
		job.ID, job.UserID, job.DockerImage)

	return c.Status(fiber.StatusCreated).JSON(response)
}

// GetJob handles GET /api/jobs/:id
func (h *JobHandler) GetJob(c *fiber.Ctx) error {
	// Parse job ID from URL params
	idParam := c.Params("id")
	jobID, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid job ID format",
			Code:    fiber.StatusBadRequest,
		})
	}

	// Retrieve job from database
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	job, err := h.jobRepo.GetJobByID(ctx, jobID)
	if err != nil {
		if err.Error() == "job not found" {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error:   "not_found",
				Message: "Job not found",
				Code:    fiber.StatusNotFound,
			})
		}

		log.Printf("Failed to get job: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve job",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.JSON(job)
}

// GetUserJobs handles GET /api/users/:userId/jobs
func (h *JobHandler) GetUserJobs(c *fiber.Ctx) error {
	userID := c.Params("userId")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "User ID is required",
			Code:    fiber.StatusBadRequest,
		})
	}

	// Get limit from query params (default: 50)
	limit := c.QueryInt("limit", 50)
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	jobs, err := h.jobRepo.GetJobsByUserID(ctx, userID, limit)
	if err != nil {
		log.Printf("Failed to get user jobs: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve jobs",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.JSON(fiber.Map{
		"user_id": userID,
		"count":   len(jobs),
		"jobs":    jobs,
	})
}
