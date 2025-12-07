package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Sambit-Mondal/karbos/server/internal/queue"
)

// PromoterService moves delayed jobs to immediate queue when scheduled time arrives
type PromoterService struct {
	queue         *queue.RedisQueue
	checkInterval time.Duration
	stopChan      chan struct{}
	doneChan      chan struct{}
}

// NewPromoterService creates a new delayed job promoter service
func NewPromoterService(queue *queue.RedisQueue, checkInterval time.Duration) *PromoterService {
	if checkInterval == 0 {
		checkInterval = 10 * time.Second // Default 10 seconds
	}
	return &PromoterService{
		queue:         queue,
		checkInterval: checkInterval,
		stopChan:      make(chan struct{}),
		doneChan:      make(chan struct{}),
	}
}

// Start begins the promoter service loop
func (p *PromoterService) Start(ctx context.Context) error {
	log.Printf("🚀 Starting delayed job promoter service (interval: %s)", p.checkInterval)

	go p.run(ctx)

	return nil
}

// Stop gracefully stops the promoter service
func (p *PromoterService) Stop() {
	log.Println("🛑 Stopping delayed job promoter service...")
	close(p.stopChan)

	// Wait for service to finish with timeout
	select {
	case <-p.doneChan:
		log.Println("✓ Delayed job promoter service stopped")
	case <-time.After(5 * time.Second):
		log.Println("⚠ Delayed job promoter service stop timeout")
	}
}

// run is the main loop that checks for ready jobs
func (p *PromoterService) run(ctx context.Context) {
	defer close(p.doneChan)

	ticker := time.NewTicker(p.checkInterval)
	defer ticker.Stop()

	log.Println("✓ Delayed job promoter service started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping promoter service")
			return
		case <-p.stopChan:
			log.Println("Stop signal received, stopping promoter service")
			return
		case <-ticker.C:
			if err := p.promoteReadyJobs(ctx); err != nil {
				log.Printf("⚠ Error promoting jobs: %v", err)
			}
		}
	}
}

// promoteReadyJobs checks delayed queue and promotes jobs whose scheduled time has arrived
func (p *PromoterService) promoteReadyJobs(ctx context.Context) error {
	// Get all jobs from delayed queue that are ready (score <= current timestamp)
	now := time.Now()
	items, err := p.queue.GetReadyDelayedJobs(ctx, now)
	if err != nil {
		return fmt.Errorf("failed to get ready delayed jobs: %w", err)
	}

	if len(items) == 0 {
		return nil // No jobs ready for promotion
	}

	log.Printf("⚡ Found %d jobs ready for promotion", len(items))

	// Promote each ready job
	promoted := 0
	failed := 0

	for _, item := range items {
		if err := p.promoteJob(ctx, item); err != nil {
			log.Printf("⚠ Failed to promote job %s: %v", item.JobID, err)
			failed++
		} else {
			promoted++
		}
	}

	log.Printf("✓ Promoted %d jobs, %d failed", promoted, failed)
	return nil
}

// promoteJob moves a single job from delayed queue to immediate queue
func (p *PromoterService) promoteJob(ctx context.Context, item *queue.QueueItem) error {
	// Add to immediate queue
	if err := p.queue.EnqueueImmediate(ctx, item); err != nil {
		return fmt.Errorf("failed to enqueue to immediate queue: %w", err)
	}

	// Remove from delayed queue
	if err := p.queue.RemoveFromDelayed(ctx, item.JobID); err != nil {
		// Log error but don't fail - job is already in immediate queue
		log.Printf("⚠ Failed to remove job %s from delayed queue: %v", item.JobID, err)
	}

	log.Printf("✓ Promoted job %s from delayed to immediate queue", item.JobID)
	return nil
}

// GetStatus returns the current status of the promoter service
func (p *PromoterService) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	// Get stats from delayed queue
	stats, err := p.queue.GetDelayedQueueStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get delayed queue stats: %w", err)
	}

	status := map[string]interface{}{
		"running":        true,
		"check_interval": p.checkInterval.String(),
		"delayed_jobs":   stats["total_delayed_jobs"],
		"ready_jobs":     stats["ready_jobs"],
	}

	return status, nil
}
