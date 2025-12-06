package worker

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/Sambit-Mondal/karbos/server/internal/database"
	"github.com/Sambit-Mondal/karbos/server/internal/docker"
	"github.com/Sambit-Mondal/karbos/server/internal/queue"
)

// Pool manages multiple worker consumers running concurrently
type Pool struct {
	consumers     []*Consumer
	size          int
	queue         *queue.RedisQueue
	jobRepo       *database.JobRepository
	executionRepo *database.ExecutionLogRepository
	dockerService *docker.Service
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

// PoolConfig holds configuration for the worker pool
type PoolConfig struct {
	Size          int
	Queue         *queue.RedisQueue
	JobRepo       *database.JobRepository
	ExecutionRepo *database.ExecutionLogRepository
	DockerService *docker.Service
}

// NewPool creates a new worker pool
func NewPool(config PoolConfig) (*Pool, error) {
	if config.Size <= 0 {
		return nil, fmt.Errorf("pool size must be greater than 0")
	}

	if config.Queue == nil {
		return nil, fmt.Errorf("queue is required")
	}

	if config.JobRepo == nil {
		return nil, fmt.Errorf("job repository is required")
	}

	if config.ExecutionRepo == nil {
		return nil, fmt.Errorf("execution repository is required")
	}

	if config.DockerService == nil {
		return nil, fmt.Errorf("docker service is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &Pool{
		size:          config.Size,
		queue:         config.Queue,
		jobRepo:       config.JobRepo,
		executionRepo: config.ExecutionRepo,
		dockerService: config.DockerService,
		consumers:     make([]*Consumer, 0, config.Size),
		ctx:           ctx,
		cancel:        cancel,
	}

	return pool, nil
}

// Start initializes and starts all worker consumers in the pool
func (p *Pool) Start() error {
	log.Printf("Starting worker pool with %d workers...", p.size)

	// Create and start each worker
	for i := 0; i < p.size; i++ {
		workerID := fmt.Sprintf("worker-%d", i+1)

		consumer := NewConsumer(
			p.queue,
			p.jobRepo,
			p.executionRepo,
			p.dockerService,
			workerID,
		)

		p.consumers = append(p.consumers, consumer)

		// Start consumer in its own goroutine
		p.wg.Add(1)
		go func(c *Consumer, id string) {
			defer p.wg.Done()
			log.Printf("[%s] Worker started", id)
			c.Start(p.ctx)
			log.Printf("[%s] Worker stopped", id)
		}(consumer, workerID)
	}

	log.Printf("Worker pool started successfully with %d workers", p.size)
	return nil
}

// Stop gracefully shuts down all workers in the pool
func (p *Pool) Stop() {
	log.Println("Stopping worker pool...")

	// Cancel context to signal all workers to stop
	p.cancel()

	// Wait for all workers to finish their current jobs
	p.wg.Wait()

	log.Println("Worker pool stopped successfully")
}

// Wait blocks until all workers have stopped
func (p *Pool) Wait() {
	p.wg.Wait()
}

// GetSize returns the number of workers in the pool
func (p *Pool) GetSize() int {
	return p.size
}

// GetConsumers returns all consumer instances (for monitoring/debugging)
func (p *Pool) GetConsumers() []*Consumer {
	return p.consumers
}

// GetStatus returns the current status of the worker pool
func (p *Pool) GetStatus() map[string]interface{} {
	workers := make([]map[string]string, len(p.consumers))
	for i, consumer := range p.consumers {
		workers[i] = map[string]string{
			"id":     consumer.GetWorkerID(),
			"status": "running",
		}
	}

	return map[string]interface{}{
		"pool_size": p.size,
		"workers":   workers,
		"status":    "active",
	}
}

// ScaleUp adds new workers to the pool (dynamic scaling)
func (p *Pool) ScaleUp(count int) error {
	if count <= 0 {
		return fmt.Errorf("scale count must be greater than 0")
	}

	log.Printf("Scaling up worker pool by %d workers...", count)

	currentSize := len(p.consumers)

	for i := 0; i < count; i++ {
		workerID := fmt.Sprintf("worker-%d", currentSize+i+1)

		consumer := NewConsumer(
			p.queue,
			p.jobRepo,
			p.executionRepo,
			p.dockerService,
			workerID,
		)

		p.consumers = append(p.consumers, consumer)
		p.size++

		// Start new consumer
		p.wg.Add(1)
		go func(c *Consumer, id string) {
			defer p.wg.Done()
			log.Printf("[%s] Worker started", id)
			c.Start(p.ctx)
			log.Printf("[%s] Worker stopped", id)
		}(consumer, workerID)
	}

	log.Printf("Scaled up to %d workers", p.size)
	return nil
}

// HealthCheck verifies that the worker pool and its dependencies are healthy
func (p *Pool) HealthCheck(ctx context.Context) error {
	// Check Redis connection
	if err := p.queue.HealthCheck(ctx); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}

	// Check Docker daemon
	if err := p.dockerService.Ping(ctx); err != nil {
		return fmt.Errorf("docker health check failed: %w", err)
	}

	// All checks passed
	return nil
}
