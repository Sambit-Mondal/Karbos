package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/Sambit-Mondal/karbos/server/internal/config"
	"github.com/Sambit-Mondal/karbos/server/internal/database"
	"github.com/Sambit-Mondal/karbos/server/internal/models"
	"github.com/google/uuid"
)

// DemoDataSeeder handles seeding demo data into the system
type DemoDataSeeder struct {
	db            *database.DB
	jobRepo       *database.JobRepository
	executionRepo *database.ExecutionLogRepository
	carbonRepo    *database.CarbonCacheRepository
	apiURL        string
	regions       []string
	dockerImages  []string
	users         []string
}

// JobSubmitRequest represents the API job submission payload
type JobSubmitRequest struct {
	UserID      string   `json:"user_id"`
	DockerImage string   `json:"docker_image"`
	Command     []string `json:"command"`
	Region      string   `json:"region"`
	Deadline    string   `json:"deadline,omitempty"`
}

func main() {
	log.Println("🌱 Starting Karbos Demo Data Seeder...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.NewDatabase(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✓ Connected to database")

	// Initialize repositories
	jobRepo := database.NewJobRepository(db)
	executionRepo := database.NewExecutionLogRepository(db.DB) // Needs *sql.DB
	carbonRepo := database.NewCarbonCacheRepository(db)

	// Create seeder
	seeder := &DemoDataSeeder{
		db:            db,
		jobRepo:       jobRepo,
		executionRepo: executionRepo,
		carbonRepo:    carbonRepo,
		apiURL:        fmt.Sprintf("http://localhost:%s/api", cfg.Server.Port),
		regions: []string{
			"US-EAST", "US-WEST", "US-CENTRAL",
			"EU-WEST", "EU-CENTRAL", "EU-NORTH",
			"ASIA-EAST", "ASIA-SOUTH", "ASIA-SOUTHEAST",
			"AU-EAST", "SA-EAST", "AF-SOUTH",
		},
		dockerImages: []string{
			"alpine:latest",
			"python:3.9-alpine",
			"python:3.11-slim",
			"node:18-alpine",
			"node:20-alpine",
			"golang:1.21-alpine",
			"ubuntu:22.04",
			"nginx:alpine",
			"redis:7-alpine",
			"postgres:15-alpine",
		},
		users: []string{
			"demo-user-analytics",
			"demo-user-ml-training",
			"demo-user-data-pipeline",
			"demo-user-batch-processing",
			"demo-user-etl-jobs",
			"demo-user-video-encoding",
			"demo-user-image-processing",
			"demo-user-report-generation",
			"demo-user-backup-jobs",
			"demo-user-testing",
		},
	}

	ctx := context.Background()

	// Seed carbon cache data
	log.Println("\n📊 Seeding carbon intensity cache...")
	if err := seeder.seedCarbonCache(ctx); err != nil {
		log.Printf("Warning: Failed to seed carbon cache: %v", err)
	} else {
		log.Println("✓ Seeded carbon intensity data for all regions")
	}

	// Seed historical jobs (last 24 hours)
	log.Println("\n📜 Seeding 50 historical jobs (last 24 hours)...")
	historicalCount, err := seeder.seedHistoricalJobs(ctx, 50)
	if err != nil {
		log.Printf("Warning: Only seeded %d historical jobs: %v", historicalCount, err)
	} else {
		log.Printf("✓ Seeded %d historical jobs", historicalCount)
	}

	// Wait a moment for API to be ready
	time.Sleep(2 * time.Second)

	// Submit active jobs via API
	log.Println("\n🚀 Submitting 5 active jobs via API...")
	activeCount, err := seeder.submitActiveJobs(5)
	if err != nil {
		log.Printf("Warning: Only submitted %d active jobs: %v", activeCount, err)
	} else {
		log.Printf("✓ Submitted %d active jobs", activeCount)
	}

	// Print summary
	seeder.printSummary(ctx)

	log.Println("\n✅ Demo data seeding complete!")
	log.Println("\n💡 Next steps:")
	log.Println("   1. Start the API server: go run cmd/api/main.go")
	log.Println("   2. Start the worker node: go run cmd/worker/main.go")
	log.Println("   3. Open your dashboard at http://localhost:3000")
	log.Println("   4. Watch the active jobs being processed!")
}

// seedCarbonCache populates carbon intensity cache for all regions
func (s *DemoDataSeeder) seedCarbonCache(ctx context.Context) error {
	now := time.Now()
	baseIntensities := map[string]float64{
		"US-EAST":        320.5,
		"US-WEST":        180.2,
		"US-CENTRAL":     420.8,
		"EU-WEST":        150.3,
		"EU-CENTRAL":     280.7,
		"EU-NORTH":       90.4,
		"ASIA-EAST":      580.9,
		"ASIA-SOUTH":     710.2,
		"ASIA-SOUTHEAST": 650.5,
		"AU-EAST":        420.3,
		"SA-EAST":        250.6,
		"AF-SOUTH":       680.1,
	}

	count := 0
	for region, baseIntensity := range baseIntensities {
		// Create cache entries for last 24 hours (hourly)
		for i := 0; i < 24; i++ {
			timestamp := now.Add(-time.Duration(i) * time.Hour)

			// Add some variation to intensity
			variation := rand.Float64()*100 - 50 // -50 to +50
			intensity := baseIntensity + variation

			// Save to cache with 2-hour TTL
			if err := s.carbonRepo.SaveCarbonIntensity(
				ctx,
				region,
				timestamp,
				intensity,
				"gCO2eq/kWh",
				2*time.Hour,
			); err != nil {
				return fmt.Errorf("failed to cache carbon data for %s: %w", region, err)
			}
			count++
		}
	}

	log.Printf("   Cached %d carbon intensity data points", count)
	return nil
}

// seedHistoricalJobs creates fake completed/failed jobs from the last 24 hours
func (s *DemoDataSeeder) seedHistoricalJobs(ctx context.Context, count int) (int, error) {
	now := time.Now()
	seeded := 0

	jobTypes := []struct {
		name        string
		command     []string
		duration    time.Duration
		probability float64 // Probability of this job type
	}{
		{"quick-echo", []string{"echo", "Task completed"}, 2 * time.Second, 0.3},
		{"data-processing", []string{"sh", "-c", "echo Processing data...; sleep 15; echo Done"}, 15 * time.Second, 0.25},
		{"ml-training", []string{"python", "-c", "import time; print('Training model...'); time.sleep(45); print('Model trained')"}, 45 * time.Second, 0.15},
		{"batch-analysis", []string{"sh", "-c", "echo Analyzing batch...; sleep 30; echo Analysis complete"}, 30 * time.Second, 0.15},
		{"report-generation", []string{"sh", "-c", "echo Generating report...; sleep 20; echo Report ready"}, 20 * time.Second, 0.10},
		{"backup-task", []string{"sh", "-c", "echo Backing up...; sleep 10; echo Backup complete"}, 10 * time.Second, 0.05},
	}

	for i := 0; i < count; i++ {
		// Random time in the last 24 hours
		hoursAgo := rand.Float64() * 24
		createdAt := now.Add(-time.Duration(hoursAgo * float64(time.Hour)))

		// Select random job type based on probability
		jobType := s.selectJobType(jobTypes)

		// Random user, region, and docker image
		userID := s.users[rand.Intn(len(s.users))]
		region := s.regions[rand.Intn(len(s.regions))]
		dockerImage := s.dockerImages[rand.Intn(len(s.dockerImages))]

		// 85% success rate, 15% failure rate
		status := models.JobStatusCompleted
		exitCode := 0
		var errorMessage *string
		if rand.Float64() < 0.15 { // 15% failure rate
			status = models.JobStatusFailed
			exitCode = rand.Intn(10) + 1 // Exit codes 1-10
			errMsg := fmt.Sprintf("Container exited with code %d", exitCode)
			errorMessage = &errMsg
		}

		// Convert command to JSON string
		cmdJSON, _ := json.Marshal(jobType.command)
		cmdStr := string(cmdJSON)

		// Create job
		scheduledTime := createdAt.Add(1 * time.Minute)
		startedAt := scheduledTime
		completedAt := startedAt.Add(jobType.duration)

		job := &models.Job{
			ID:            uuid.New(),
			UserID:        userID,
			DockerImage:   dockerImage,
			Command:       &cmdStr,
			Status:        status,
			Region:        &region,
			ScheduledTime: &scheduledTime,
			Deadline:      createdAt.Add(4 * time.Hour),
			CreatedAt:     createdAt,
			StartedAt:     &startedAt,
			CompletedAt:   &completedAt,
		}

		// Insert job
		if err := s.jobRepo.CreateJob(ctx, job); err != nil {
			log.Printf("Warning: Failed to create job %d: %v", i+1, err)
			continue
		}

		// Create execution log
		output := fmt.Sprintf("Job %s executed successfully\n%s", jobType.name, jobType.command[len(jobType.command)-1])
		if status == models.JobStatusFailed {
			output = fmt.Sprintf("Job %s failed\nError: %s", jobType.name, *errorMessage)
		}

		durationSeconds := int(jobType.duration.Seconds())

		executionLog := &models.ExecutionLog{
			ID:           uuid.New(),
			JobID:        job.ID,
			StartedAt:    startedAt,
			CompletedAt:  &completedAt,
			ExitCode:     exitCode,
			Output:       output,
			ErrorMessage: errorMessage,
			Duration:     durationSeconds,
			CreatedAt:    completedAt,
		}

		if err := s.executionRepo.CreateExecutionLog(ctx, executionLog); err != nil {
			log.Printf("Warning: Failed to create execution log for job %s: %v", job.ID, err)
			continue
		}

		seeded++

		// Show progress every 10 jobs
		if (i+1)%10 == 0 {
			log.Printf("   Progress: %d/%d jobs seeded", i+1, count)
		}
	}

	return seeded, nil
}

// selectJobType randomly selects a job type based on probability
func (s *DemoDataSeeder) selectJobType(types []struct {
	name        string
	command     []string
	duration    time.Duration
	probability float64
}) struct {
	name        string
	command     []string
	duration    time.Duration
	probability float64
} {
	r := rand.Float64()
	cumulative := 0.0

	for _, jt := range types {
		cumulative += jt.probability
		if r <= cumulative {
			return jt
		}
	}

	// Fallback to first type
	return types[0]
}

// submitActiveJobs submits real jobs via the API
func (s *DemoDataSeeder) submitActiveJobs(count int) (int, error) {
	submitted := 0

	activeJobs := []struct {
		name     string
		request  JobSubmitRequest
		deadline bool
	}{
		{
			name: "Real-time Analytics Pipeline",
			request: JobSubmitRequest{
				UserID:      "demo-user-analytics",
				DockerImage: "python:3.9-alpine",
				Command:     []string{"python", "-c", "import time; print('Running analytics...'); time.sleep(30); print('Analytics complete')"},
				Region:      "US-WEST",
			},
			deadline: false,
		},
		{
			name: "ML Model Training Job",
			request: JobSubmitRequest{
				UserID:      "demo-user-ml-training",
				DockerImage: "python:3.11-slim",
				Command:     []string{"python", "-c", "import time; print('Training ML model...'); time.sleep(45); print('Model trained and saved')"},
				Region:      "EU-WEST",
			},
			deadline: true,
		},
		{
			name: "Video Transcoding Task",
			request: JobSubmitRequest{
				UserID:      "demo-user-video-encoding",
				DockerImage: "alpine:latest",
				Command:     []string{"sh", "-c", "echo 'Transcoding video...'; sleep 60; echo 'Video transcoded successfully'"},
				Region:      "ASIA-EAST",
			},
			deadline: true,
		},
		{
			name: "Database Backup Job",
			request: JobSubmitRequest{
				UserID:      "demo-user-backup-jobs",
				DockerImage: "postgres:15-alpine",
				Command:     []string{"sh", "-c", "echo 'Starting backup...'; sleep 40; echo 'Backup completed'"},
				Region:      "US-EAST",
			},
			deadline: false,
		},
		{
			name: "Report Generation Task",
			request: JobSubmitRequest{
				UserID:      "demo-user-report-generation",
				DockerImage: "node:20-alpine",
				Command:     []string{"sh", "-c", "echo 'Generating monthly report...'; sleep 25; echo 'Report generated'"},
				Region:      "EU-CENTRAL",
			},
			deadline: true,
		},
	}

	for i := 0; i < count && i < len(activeJobs); i++ {
		job := activeJobs[i]

		// Add deadline if specified (4 hours from now)
		if job.deadline {
			deadline := time.Now().Add(4 * time.Hour)
			job.request.Deadline = deadline.Format(time.RFC3339)
		}

		// Submit job to API
		if err := s.submitJobToAPI(job.request); err != nil {
			log.Printf("Warning: Failed to submit '%s': %v", job.name, err)
			continue
		}

		log.Printf("   ✓ Submitted: %s", job.name)
		submitted++

		// Small delay between submissions
		time.Sleep(500 * time.Millisecond)
	}

	return submitted, nil
}

// submitJobToAPI sends a job submission request to the API
func (s *DemoDataSeeder) submitJobToAPI(req JobSubmitRequest) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/submit", s.apiURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to submit job: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return nil
}

// printSummary displays a summary of the seeded data
func (s *DemoDataSeeder) printSummary(ctx context.Context) {
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("📊 DEMO DATA SUMMARY")
	log.Println(strings.Repeat("=", 60))

	// Count jobs by status
	var totalJobs, completedJobs, failedJobs, pendingJobs int

	rows, err := s.db.QueryContext(ctx, `
		SELECT status, COUNT(*) as count
		FROM jobs
		GROUP BY status
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int
			if err := rows.Scan(&status, &count); err == nil {
				totalJobs += count
				switch status {
				case "completed":
					completedJobs = count
				case "failed":
					failedJobs = count
				case "pending":
					pendingJobs = count
				}
			}
		}
	}

	// Count carbon cache entries
	var carbonCacheCount int
	s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM carbon_cache
	`).Scan(&carbonCacheCount)

	// Print statistics
	log.Printf("\n📈 Job Statistics:")
	log.Printf("   Total Jobs:      %d", totalJobs)
	if totalJobs > 0 {
		log.Printf("   ✓ Completed:     %d (%.1f%%)", completedJobs, float64(completedJobs)/float64(totalJobs)*100)
		log.Printf("   ✗ Failed:        %d (%.1f%%)", failedJobs, float64(failedJobs)/float64(totalJobs)*100)
		log.Printf("   ⏳ Pending:       %d (%.1f%%)", pendingJobs, float64(pendingJobs)/float64(totalJobs)*100)
	}

	log.Printf("\n💾 Cache Statistics:")
	log.Printf("   Carbon Cache:    %d entries", carbonCacheCount)

	log.Printf("\n🎯 Active Jobs:")
	log.Printf("   Pending/Running: %d jobs ready for workers", pendingJobs)

	log.Println("\n" + strings.Repeat("=", 60))
}
