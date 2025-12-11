package handlers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Sambit-Mondal/karbos/server/internal/database"
	"github.com/Sambit-Mondal/karbos/server/internal/models"
	"github.com/Sambit-Mondal/karbos/server/internal/queue"
	"github.com/Sambit-Mondal/karbos/server/internal/scheduler"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// JobHandler handles job-related HTTP requests
type JobHandler struct {
	jobRepo   *database.JobRepository
	queue     *queue.RedisQueue
	scheduler *scheduler.CarbonScheduler
}

// NewJobHandler creates a new job handler
func NewJobHandler(jobRepo *database.JobRepository, queue *queue.RedisQueue, scheduler *scheduler.CarbonScheduler) *JobHandler {
	return &JobHandler{
		jobRepo:   jobRepo,
		queue:     queue,
		scheduler: scheduler,
	}
}

// SubmitJob handles POST /api/submit
func (h *JobHandler) SubmitJob(c *fiber.Ctx) error {
	var req models.SubmitJobRequest

	// Check for dry-run mode
	dryRun := c.Query("dry_run") == "true"

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

	// Set default region if not provided
	region := "US-EAST" // Default region
	if req.Region != nil && *req.Region != "" {
		region = *req.Region
	}

	// Determine estimated duration
	var estimatedDuration time.Duration
	if req.EstimatedDuration != nil && *req.EstimatedDuration > 0 {
		estimatedDuration = time.Duration(*req.EstimatedDuration) * time.Second
	} else {
		estimatedDuration = 10 * time.Minute // Default 10 minutes
	}

	// Carbon-aware scheduling
	var scheduledTime time.Time
	var immediate bool = true
	var expectedIntensity float64 = 0
	var carbonSavings float64 = 0

	// Create context for scheduling
	schedCtx, schedCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer schedCancel()

	if h.scheduler != nil {
		// Create scheduling request
		schedReq := &scheduler.ScheduleRequest{
			Region:     region,
			Duration:   estimatedDuration,
			Deadline:   deadline,
			WindowSize: 24 * time.Hour,
		}

		// Get scheduling recommendation
		schedResult, err := h.scheduler.Schedule(schedCtx, schedReq)
		if err != nil {
			log.Printf("⚠ Scheduling failed, defaulting to immediate: %v", err)
			// Continue with immediate execution
		} else {
			scheduledTime = schedResult.ScheduledTime
			immediate = schedResult.Immediate
			expectedIntensity = schedResult.ExpectedIntensity
			carbonSavings = schedResult.CarbonSavings

			log.Printf("✓ Carbon scheduling: immediate=%v, scheduled=%v, savings=%.2f gCO2eq/kWh",
				immediate, scheduledTime.Format(time.RFC3339), carbonSavings)
		}
	}

	// If no scheduled time determined, use now
	if scheduledTime.IsZero() {
		scheduledTime = time.Now()
	}

	// Serialize command array to JSON string for database storage
	var commandStr *string
	if len(req.Command) > 0 {
		cmdJSON, err := json.Marshal(req.Command)
		if err != nil {
			log.Printf("Failed to serialize command: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Error:   "invalid_command",
				Message: "Failed to process command",
				Code:    fiber.StatusBadRequest,
			})
		}
		cmdJSONStr := string(cmdJSON)
		commandStr = &cmdJSONStr
	}

	// Create job object
	job := &models.Job{
		ID:                uuid.New(),
		UserID:            req.UserID,
		DockerImage:       req.DockerImage,
		Command:           commandStr,
		Status:            models.JobStatusPending,
		Deadline:          deadline,
		EstimatedDuration: req.EstimatedDuration,
		Region:            &region,
		ScheduledTime:     &scheduledTime,
		CreatedAt:         time.Now(),
		Metadata:          "{}",
	}

	// If dry-run mode, return prediction without saving
	if dryRun {
		response := models.SubmitJobResponse{
			JobID:             job.ID.String(),
			Status:            models.JobStatusPending,
			CreatedAt:         job.CreatedAt,
			ScheduledTime:     scheduledTime.Format(time.RFC3339),
			Immediate:         immediate,
			ExpectedIntensity: expectedIntensity,
			CarbonSavings:     carbonSavings,
			Message:           "Dry run - job not created",
		}

		log.Printf("✓ Dry run completed: immediate=%v, savings=%.2f gCO2eq/kWh", immediate, carbonSavings)
		return c.JSON(response)
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
		ScheduledTime: scheduledTime,
		Priority:      0,
	}

	// Route to appropriate queue based on scheduling decision
	if immediate {
		// Push to Redis immediate queue (FIFO List)
		if err := h.queue.EnqueueImmediate(ctx, queueItem); err != nil {
			log.Printf("Failed to enqueue immediate job: %v", err)
		} else {
			log.Printf("✓ Job queued for immediate execution: %s", job.ID)
		}
	} else {
		// Push to Redis delayed queue (Sorted Set with scheduled_time as score)
		if err := h.queue.EnqueueDelayed(ctx, queueItem); err != nil {
			log.Printf("Failed to enqueue delayed job: %v", err)
		} else {
			log.Printf("✓ Job scheduled for later execution at %s: %s",
				scheduledTime.Format(time.RFC3339), job.ID)
		}
	}

	// Prepare response
	response := models.SubmitJobResponse{
		JobID:             job.ID.String(),
		Status:            job.Status,
		CreatedAt:         job.CreatedAt,
		ScheduledTime:     scheduledTime.Format(time.RFC3339),
		Immediate:         immediate,
		ExpectedIntensity: expectedIntensity,
		CarbonSavings:     carbonSavings,
		Message:           "Job submitted successfully",
	}

	if !immediate {
		response.Message = "Job scheduled for optimal carbon efficiency"
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

// GetAllJobs handles GET /api/jobs
func (h *JobHandler) GetAllJobs(c *fiber.Ctx) error {
	// Get limit from query params (default: 100)
	limit := c.QueryInt("limit", 100)
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get all jobs
	jobs, err := h.jobRepo.GetAllJobs(ctx, limit)
	if err != nil {
		log.Printf("Failed to get all jobs: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve jobs",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.JSON(jobs)
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
