package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Sambit-Mondal/karbos/server/internal/models"

	"github.com/google/uuid"
)

// ExecutionLogRepository handles execution log operations
type ExecutionLogRepository struct {
	db *sql.DB
}

// NewExecutionLogRepository creates a new execution log repository
func NewExecutionLogRepository(db *sql.DB) *ExecutionLogRepository {
	return &ExecutionLogRepository{db: db}
}

// CreateExecutionLog creates a new execution log entry
func (r *ExecutionLogRepository) CreateExecutionLog(ctx context.Context, log *models.ExecutionLog) error {
	query := `
		INSERT INTO execution_logs (
			id, job_id, output, error_message, exit_code, 
			duration, started_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Generate UUID if not provided
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		log.ID,
		log.JobID,
		log.Output,
		log.ErrorMessage,
		log.ExitCode,
		log.Duration,
		log.StartedAt,
		log.CompletedAt,
	).Scan(&log.ID, &log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create execution log: %w", err)
	}

	return nil
}

// GetExecutionLogByJobID retrieves the execution log for a specific job
func (r *ExecutionLogRepository) GetExecutionLogByJobID(ctx context.Context, jobID uuid.UUID) (*models.ExecutionLog, error) {
	query := `
		SELECT 
			id, job_id, output, error_message, exit_code, 
			duration, started_at, completed_at, created_at
		FROM execution_logs
		WHERE job_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	log := &models.ExecutionLog{}
	var errorMessage sql.NullString
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, jobID).Scan(
		&log.ID,
		&log.JobID,
		&log.Output,
		&errorMessage,
		&log.ExitCode,
		&log.Duration,
		&log.StartedAt,
		&completedAt,
		&log.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("execution log not found for job %s", jobID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get execution log: %w", err)
	}

	// Handle nullable fields
	if errorMessage.Valid {
		log.ErrorMessage = &errorMessage.String
	}
	if completedAt.Valid {
		log.CompletedAt = &completedAt.Time
	}

	return log, nil
}

// GetAllExecutionLogsByJobID retrieves all execution logs for a job (in case of retries)
func (r *ExecutionLogRepository) GetAllExecutionLogsByJobID(ctx context.Context, jobID uuid.UUID) ([]*models.ExecutionLog, error) {
	query := `
		SELECT 
			id, job_id, output, error_message, exit_code, 
			duration, started_at, completed_at, created_at
		FROM execution_logs
		WHERE job_id = $1
		ORDER BY created_at DESC
	`

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to query execution logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.ExecutionLog

	for rows.Next() {
		log := &models.ExecutionLog{}
		var errorMessage sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&log.ID,
			&log.JobID,
			&log.Output,
			&errorMessage,
			&log.ExitCode,
			&log.Duration,
			&log.StartedAt,
			&completedAt,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan execution log: %w", err)
		}

		// Handle nullable fields
		if errorMessage.Valid {
			log.ErrorMessage = &errorMessage.String
		}
		if completedAt.Valid {
			log.CompletedAt = &completedAt.Time
		}

		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating execution logs: %w", err)
	}

	return logs, nil
}

// UpdateExecutionLog updates an existing execution log (for streaming updates)
func (r *ExecutionLogRepository) UpdateExecutionLog(ctx context.Context, log *models.ExecutionLog) error {
	query := `
		UPDATE execution_logs
		SET 
			output = $1,
			error_message = $2,
			exit_code = $3,
			duration = $4,
			completed_at = $5
		WHERE id = $6
	`

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(
		ctx,
		query,
		log.Output,
		log.ErrorMessage,
		log.ExitCode,
		log.Duration,
		log.CompletedAt,
		log.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update execution log: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("execution log not found: %s", log.ID)
	}

	return nil
}

// DeleteExecutionLogsByJobID deletes all execution logs for a job
func (r *ExecutionLogRepository) DeleteExecutionLogsByJobID(ctx context.Context, jobID uuid.UUID) error {
	query := `DELETE FROM execution_logs WHERE job_id = $1`

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to delete execution logs: %w", err)
	}

	return nil
}

// GetRecentExecutionLogs retrieves the most recent execution logs (for monitoring)
func (r *ExecutionLogRepository) GetRecentExecutionLogs(ctx context.Context, limit int) ([]*models.ExecutionLog, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT 
			id, job_id, output, error_message, exit_code, 
			duration, started_at, completed_at, created_at
		FROM execution_logs
		ORDER BY created_at DESC
		LIMIT $1
	`

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent execution logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.ExecutionLog

	for rows.Next() {
		log := &models.ExecutionLog{}
		var errorMessage sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&log.ID,
			&log.JobID,
			&log.Output,
			&errorMessage,
			&log.ExitCode,
			&log.Duration,
			&log.StartedAt,
			&completedAt,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan execution log: %w", err)
		}

		// Handle nullable fields
		if errorMessage.Valid {
			log.ErrorMessage = &errorMessage.String
		}
		if completedAt.Valid {
			log.CompletedAt = &completedAt.Time
		}

		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating execution logs: %w", err)
	}

	return logs, nil
}
