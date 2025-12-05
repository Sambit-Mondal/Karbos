package models

import (
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "PENDING"
	JobStatusDelayed   JobStatus = "DELAYED"
	JobStatusRunning   JobStatus = "RUNNING"
	JobStatusCompleted JobStatus = "COMPLETED"
	JobStatusFailed    JobStatus = "FAILED"
)

// Job represents a job submission in the system
type Job struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	UserID            string     `json:"user_id" db:"user_id"`
	DockerImage       string     `json:"docker_image" db:"docker_image"`
	Command           *string    `json:"command,omitempty" db:"command"`
	Status            JobStatus  `json:"status" db:"status"`
	ScheduledTime     *time.Time `json:"scheduled_time,omitempty" db:"scheduled_time"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	StartedAt         *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	Deadline          time.Time  `json:"deadline" db:"deadline"`
	EstimatedDuration *int       `json:"estimated_duration,omitempty" db:"estimated_duration"` // in seconds
	Region            *string    `json:"region,omitempty" db:"region"`
	Metadata          string     `json:"metadata,omitempty" db:"metadata"` // JSON stored as string
}

// ExecutionLog represents a log entry for job execution
type ExecutionLog struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	JobID        uuid.UUID  `json:"job_id" db:"job_id"`
	Output       *string    `json:"output,omitempty" db:"output"`
	ErrorOutput  *string    `json:"error_output,omitempty" db:"error_output"`
	ExitCode     *int       `json:"exit_code,omitempty" db:"exit_code"`
	Duration     *int       `json:"duration,omitempty" db:"duration"` // in seconds
	StartedAt    time.Time  `json:"started_at" db:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	WorkerNodeID *string    `json:"worker_node_id,omitempty" db:"worker_node_id"`
}

// CarbonCache represents cached carbon intensity data
type CarbonCache struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Region         string    `json:"region" db:"region"`
	Timestamp      time.Time `json:"timestamp" db:"timestamp"`
	IntensityValue float64   `json:"intensity_value" db:"intensity_value"` // gCO2/kWh
	ForecastWindow *int      `json:"forecast_window,omitempty" db:"forecast_window"`
	Source         *string   `json:"source,omitempty" db:"source"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// SubmitJobRequest represents the API request for job submission
type SubmitJobRequest struct {
	UserID            string  `json:"user_id" validate:"required"`
	DockerImage       string  `json:"docker_image" validate:"required"`
	Command           *string `json:"command,omitempty"`
	Deadline          string  `json:"deadline" validate:"required"` // ISO 8601 format
	EstimatedDuration *int    `json:"estimated_duration,omitempty"` // in seconds
	Region            *string `json:"region,omitempty"`
}

// SubmitJobResponse represents the API response for job submission
type SubmitJobResponse struct {
	JobID     string    `json:"job_id"`
	Status    JobStatus `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// ValidateStatus checks if the status is valid
func (s JobStatus) IsValid() bool {
	switch s {
	case JobStatusPending, JobStatusDelayed, JobStatusRunning, JobStatusCompleted, JobStatusFailed:
		return true
	}
	return false
}
