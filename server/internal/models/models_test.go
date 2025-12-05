package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJobStatus_IsValid(t *testing.T) {
	tests := []struct {
		status JobStatus
		want   bool
	}{
		{JobStatusPending, true},
		{JobStatusDelayed, true},
		{JobStatusRunning, true},
		{JobStatusCompleted, true},
		{JobStatusFailed, true},
		{JobStatus("INVALID"), false},
		{JobStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("JobStatus.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJob_Creation(t *testing.T) {
	job := &Job{
		ID:          uuid.New(),
		UserID:      "test-user",
		DockerImage: "python:3.11",
		Status:      JobStatusPending,
		Deadline:    time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
		Metadata:    "{}",
	}

	if job.ID == uuid.Nil {
		t.Error("Job ID should not be nil")
	}

	if job.Status != JobStatusPending {
		t.Errorf("Expected status PENDING, got %s", job.Status)
	}

	if job.Deadline.Before(time.Now()) {
		t.Error("Deadline should be in the future")
	}
}

func TestSubmitJobRequest_RequiredFields(t *testing.T) {
	req := SubmitJobRequest{
		UserID:      "test-user",
		DockerImage: "python:3.11",
		Deadline:    time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	if req.UserID == "" {
		t.Error("UserID is required")
	}

	if req.DockerImage == "" {
		t.Error("DockerImage is required")
	}

	if req.Deadline == "" {
		t.Error("Deadline is required")
	}
}
