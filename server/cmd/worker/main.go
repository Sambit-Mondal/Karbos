package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sambit-Mondal/karbos/server/internal/config"
	"github.com/Sambit-Mondal/karbos/server/internal/database"
	"github.com/Sambit-Mondal/karbos/server/internal/docker"
	"github.com/Sambit-Mondal/karbos/server/internal/queue"
	"github.com/Sambit-Mondal/karbos/server/internal/worker"
	"github.com/google/uuid"
)

func main() {
	log.Println("=== Karbos Worker Node Starting ===")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Environment: %s", cfg.Server.Environment)
	log.Printf("Worker Pool Size: %d", cfg.Worker.PoolSize)

	// Initialize database connection
	log.Println("Connecting to database...")
	db, err := database.NewDatabase(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	// Initialize Redis queue
	log.Println("Connecting to Redis...")
	redisAddr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	redisQueue, err := queue.NewRedisQueue(
		redisAddr,
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Queue.ImmediateQueueKey,
		cfg.Queue.DelayedSetKey,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Redis queue: %v", err)
	}
	defer redisQueue.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := redisQueue.HealthCheck(ctx); err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}
	log.Println("Redis connected successfully")

	// Initialize Docker service
	log.Println("Connecting to Docker daemon...")
	dockerService, err := docker.NewDockerService()
	if err != nil {
		log.Fatalf("Failed to initialize Docker service: %v", err)
	}
	defer dockerService.Close()

	// Test Docker connection
	if err := dockerService.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping Docker daemon: %v", err)
	}
	log.Println("Docker daemon connected successfully")

	// Get Docker info
	dockerInfo, err := dockerService.GetDockerInfo(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get Docker info: %v", err)
	} else {
		log.Printf("Docker Server Version: %v", dockerInfo["server_version"])
		log.Printf("Docker CPUs: %v", dockerInfo["cpus"])
	}

	// Initialize repositories
	jobRepo := database.NewJobRepository(db)
	executionRepo := database.NewExecutionLogRepository(db.DB)

	// Create worker pool
	log.Printf("Creating worker pool with %d workers...", cfg.Worker.PoolSize)
	workerPool, err := worker.NewPool(worker.PoolConfig{
		Size:          cfg.Worker.PoolSize,
		Queue:         redisQueue,
		JobRepo:       jobRepo,
		ExecutionRepo: executionRepo,
		DockerService: dockerService,
	})
	if err != nil {
		log.Fatalf("Failed to create worker pool: %v", err)
	}

	// Start worker pool
	if err := workerPool.Start(); err != nil {
		log.Fatalf("Failed to start worker pool: %v", err)
	}

	// Generate unique worker ID
	workerID := uuid.New().String()
	log.Printf("Worker ID: %s", workerID)

	// Start heartbeat goroutine
	heartbeatCtx, heartbeatCancel := context.WithCancel(context.Background())
	defer heartbeatCancel()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		// Send initial heartbeat
		if err := redisQueue.SetWorkerHeartbeat(heartbeatCtx, workerID, 15); err != nil {
			log.Printf("Failed to send initial heartbeat: %v", err)
		}

		for {
			select {
			case <-ticker.C:
				if err := redisQueue.SetWorkerHeartbeat(heartbeatCtx, workerID, 15); err != nil {
					log.Printf("Failed to send heartbeat: %v", err)
				} else {
					log.Printf("💓 Heartbeat sent (worker:%s)", workerID)
				}
			case <-heartbeatCtx.Done():
				log.Println("Heartbeat stopped")
				return
			}
		}
	}()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	log.Println("=== Worker Node Running ===")
	log.Println("Press Ctrl+C to stop...")

	// Wait for shutdown signal
	<-sigChan
	log.Println("\n=== Shutdown signal received ===")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop worker pool gracefully
	log.Println("Stopping worker pool...")
	workerPool.Stop()

	// Wait for all workers to finish with timeout
	done := make(chan struct{})
	go func() {
		workerPool.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Worker pool stopped gracefully")
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout reached, forcing exit")
	}

	log.Println("=== Worker Node Stopped ===")
}
