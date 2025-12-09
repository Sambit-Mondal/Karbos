package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Sambit-Mondal/karbos/server/internal/database"
	"github.com/Sambit-Mondal/karbos/server/internal/docker"
	"github.com/Sambit-Mondal/karbos/server/internal/models"
	"github.com/Sambit-Mondal/karbos/server/internal/queue"

	"github.com/google/uuid"
)

// Consumer handles job processing from Redis queue
type Consumer struct {
	queue         *queue.RedisQueue
	jobRepo       *database.JobRepository
	executionRepo *database.ExecutionLogRepository
	dockerService *docker.Service
	pool          *Pool // Reference to parent pool for job tracking
	stopCh        chan struct{}
	workerID      string
	pollInterval  time.Duration
	jobTimeout    time.Duration
}

// NewConsumer creates a new worker consumer
func NewConsumer(
	queue *queue.RedisQueue,
	jobRepo *database.JobRepository,
	executionRepo *database.ExecutionLogRepository,
	dockerService *docker.Service,
	workerID string,
) *Consumer {
	return &Consumer{
		queue:         queue,
		jobRepo:       jobRepo,
		executionRepo: executionRepo,
		dockerService: dockerService,
		pool:          nil, // Will be set by pool after creation
		stopCh:        make(chan struct{}),
		workerID:      workerID,
		pollInterval:  2 * time.Second,  // Poll every 2 seconds
		jobTimeout:    10 * time.Minute, // 10 minute timeout per job
	}
}

// SetPool sets the parent pool reference (called by pool after consumer creation)
func (c *Consumer) SetPool(pool *Pool) {
	c.pool = pool
}

// Start begins the consumer polling loop
func (c *Consumer) Start(ctx context.Context) {
	log.Printf("[Worker %s] Starting consumer...", c.workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Worker %s] Context cancelled, stopping consumer", c.workerID)
			return
		case <-c.stopCh:
			log.Printf("[Worker %s] Stop signal received, stopping consumer", c.workerID)
			return
		default:
			// Try to dequeue and process a job
			if err := c.processNextJob(ctx); err != nil {
				// Log error but continue polling
				if err.Error() != "no jobs available" {
					log.Printf("[Worker %s] Error processing job: %v", c.workerID, err)
				}
			}

			// Sleep before next poll
			time.Sleep(c.pollInterval)
		}
	}
}

// Stop gracefully stops the consumer
func (c *Consumer) Stop() {
	close(c.stopCh)
}

// processNextJob attempts to dequeue and process one job
func (c *Consumer) processNextJob(ctx context.Context) error {
	// Check if pool is draining - stop accepting new jobs
	if c.pool != nil && c.pool.IsDraining() {
		return fmt.Errorf("worker pool is draining, not accepting new jobs")
	}

	// Dequeue from Redis
	queueItem, err := c.queue.DequeueImmediate(ctx)
	if err != nil {
		return fmt.Errorf("failed to dequeue job: %w", err)
	}

	// Check if queue is empty
	if queueItem == nil {
		return fmt.Errorf("no jobs available")
	}

	jobID, err := uuid.Parse(queueItem.JobID)
	if err != nil {
		return fmt.Errorf("invalid job ID: %w", err)
	}

	log.Printf("[Worker %s] Processing job: %s", c.workerID, jobID)

	// Process the job
	return c.executeJob(ctx, jobID)
}

// executeJob runs the complete job lifecycle
func (c *Consumer) executeJob(ctx context.Context, jobID uuid.UUID) error {
	// Create job-specific context with timeout
	jobCtx, cancel := context.WithTimeout(ctx, c.jobTimeout)
	defer cancel()

	// Fetch job details from database
	job, err := c.jobRepo.GetJobByID(jobCtx, jobID)
	if err != nil {
		return fmt.Errorf("failed to fetch job: %w", err)
	}

	// Update status to RUNNING
	job.Status = models.JobStatusRunning
	if err := c.jobRepo.UpdateJobStatus(jobCtx, jobID, models.JobStatusRunning); err != nil {
		return fmt.Errorf("failed to update job status to RUNNING: %w", err)
	}

	log.Printf("[Worker %s] Job %s: Status updated to RUNNING", c.workerID, jobID)

	// Track job start if pool is available
	jobIDStr := jobID.String()
	if c.pool != nil {
		c.pool.TrackJobStart(jobIDStr)
		defer c.pool.TrackJobComplete(jobIDStr)
	}

	// Execute Docker container
	startTime := time.Now()
	result, err := c.dockerService.RunContainer(jobCtx, job.DockerImage, nil)

	// Prepare execution log
	executionLog := &models.ExecutionLog{
		ID:        uuid.New(),
		JobID:     jobID,
		StartedAt: startTime,
		ExitCode:  result.ExitCode,
		Duration:  result.Duration,
	}

	// Handle execution result
	var finalStatus models.JobStatus
	if err != nil || result.Error != nil {
		// Job failed
		finalStatus = models.JobStatusFailed
		errorMsg := ""
		if err != nil {
			errorMsg = err.Error()
		} else if result.Error != nil {
			errorMsg = result.Error.Error()
		}
		executionLog.ErrorMessage = &errorMsg
		executionLog.Output = result.Output

		log.Printf("[Worker %s] Job %s: FAILED - %s", c.workerID, jobID, errorMsg)
	} else if result.ExitCode != 0 {
		// Container ran but exited with non-zero code
		finalStatus = models.JobStatusFailed
		errorMsg := fmt.Sprintf("Container exited with code %d", result.ExitCode)
		executionLog.ErrorMessage = &errorMsg
		executionLog.Output = result.Output

		log.Printf("[Worker %s] Job %s: FAILED - Exit code %d", c.workerID, jobID, result.ExitCode)
	} else {
		// Success
		finalStatus = models.JobStatusCompleted
		executionLog.Output = result.Output

		log.Printf("[Worker %s] Job %s: COMPLETED successfully", c.workerID, jobID)
	}

	// Set completion time
	now := time.Now()
	executionLog.CompletedAt = &now

	// Save execution log to database
	if err := c.executionRepo.CreateExecutionLog(jobCtx, executionLog); err != nil {
		log.Printf("[Worker %s] Warning: Failed to save execution log for job %s: %v", c.workerID, jobID, err)
	}

	// Update final job status
	if err := c.jobRepo.UpdateJobStatus(jobCtx, jobID, finalStatus); err != nil {
		return fmt.Errorf("failed to update final job status: %w", err)
	}

	log.Printf("[Worker %s] Job %s: Final status set to %s", c.workerID, jobID, finalStatus)

	return nil
}

// GetWorkerID returns the unique identifier for this worker
func (c *Consumer) GetWorkerID() string {
	return c.workerID
}

// SetPollInterval updates the polling interval
func (c *Consumer) SetPollInterval(interval time.Duration) {
	c.pollInterval = interval
}

// SetJobTimeout updates the job execution timeout
func (c *Consumer) SetJobTimeout(timeout time.Duration) {
	c.jobTimeout = timeout
}
