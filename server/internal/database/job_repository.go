package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Sambit-Mondal/karbos/server/internal/models"
	"github.com/google/uuid"
)

// JobRepository handles job-related database operations
type JobRepository struct {
	db *DB
}

// NewJobRepository creates a new job repository
func NewJobRepository(db *DB) *JobRepository {
	return &JobRepository{db: db}
}

// CreateJob inserts a new job into the database
func (r *JobRepository) CreateJob(ctx context.Context, job *models.Job) error {
	query := `
		INSERT INTO jobs (
			id, user_id, docker_image, command, status, 
			deadline, estimated_duration, region, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`

	// Generate UUID if not provided
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}

	// Set default status if not provided
	if job.Status == "" {
		job.Status = models.JobStatusPending
	}

	// Set created_at to now if not provided
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}

	// Set default metadata if empty
	if job.Metadata == "" {
		job.Metadata = "{}"
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		job.ID,
		job.UserID,
		job.DockerImage,
		job.Command,
		job.Status,
		job.Deadline,
		job.EstimatedDuration,
		job.Region,
		job.Metadata,
		job.CreatedAt,
	).Scan(&job.ID, &job.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	return nil
}

// GetJobByID retrieves a job by its ID
func (r *JobRepository) GetJobByID(ctx context.Context, id uuid.UUID) (*models.Job, error) {
	query := `
		SELECT 
			id, user_id, docker_image, command, status, scheduled_time,
			created_at, started_at, completed_at, deadline, 
			estimated_duration, region, metadata
		FROM jobs
		WHERE id = $1
	`

	job := &models.Job{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.UserID,
		&job.DockerImage,
		&job.Command,
		&job.Status,
		&job.ScheduledTime,
		&job.CreatedAt,
		&job.StartedAt,
		&job.CompletedAt,
		&job.Deadline,
		&job.EstimatedDuration,
		&job.Region,
		&job.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return job, nil
}

// UpdateJobStatus updates the status of a job
func (r *JobRepository) UpdateJobStatus(ctx context.Context, id uuid.UUID, status models.JobStatus) error {
	query := `
		UPDATE jobs
		SET status = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

// GetJobsByStatus retrieves jobs by status
func (r *JobRepository) GetJobsByStatus(ctx context.Context, status models.JobStatus, limit int) ([]*models.Job, error) {
	query := `
		SELECT 
			id, user_id, docker_image, command, status, scheduled_time,
			created_at, started_at, completed_at, deadline, 
			estimated_duration, region, metadata
		FROM jobs
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by status: %w", err)
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		job := &models.Job{}
		err := rows.Scan(
			&job.ID,
			&job.UserID,
			&job.DockerImage,
			&job.Command,
			&job.Status,
			&job.ScheduledTime,
			&job.CreatedAt,
			&job.StartedAt,
			&job.CompletedAt,
			&job.Deadline,
			&job.EstimatedDuration,
			&job.Region,
			&job.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating jobs: %w", err)
	}

	return jobs, nil
}

// GetAllJobs retrieves all jobs with optional limit
func (r *JobRepository) GetAllJobs(ctx context.Context, limit int) ([]*models.Job, error) {
	query := `
		SELECT 
			id, user_id, docker_image, command, status, scheduled_time,
			created_at, started_at, completed_at, deadline, 
			estimated_duration, region, metadata
		FROM jobs
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get all jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		job := &models.Job{}
		err := rows.Scan(
			&job.ID,
			&job.UserID,
			&job.DockerImage,
			&job.Command,
			&job.Status,
			&job.ScheduledTime,
			&job.CreatedAt,
			&job.StartedAt,
			&job.CompletedAt,
			&job.Deadline,
			&job.EstimatedDuration,
			&job.Region,
			&job.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating jobs: %w", err)
	}

	return jobs, nil
}

// GetJobsByUserID retrieves jobs by user ID
func (r *JobRepository) GetJobsByUserID(ctx context.Context, userID string, limit int) ([]*models.Job, error) {
	query := `
		SELECT 
			id, user_id, docker_image, command, status, scheduled_time,
			created_at, started_at, completed_at, deadline, 
			estimated_duration, region, metadata
		FROM jobs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by user: %w", err)
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		job := &models.Job{}
		err := rows.Scan(
			&job.ID,
			&job.UserID,
			&job.DockerImage,
			&job.Command,
			&job.Status,
			&job.ScheduledTime,
			&job.CreatedAt,
			&job.StartedAt,
			&job.CompletedAt,
			&job.Deadline,
			&job.EstimatedDuration,
			&job.Region,
			&job.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating jobs: %w", err)
	}

	return jobs, nil
}
