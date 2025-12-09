package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sambit-Mondal/karbos/server/internal/carbon"
	"github.com/Sambit-Mondal/karbos/server/internal/config"
	"github.com/Sambit-Mondal/karbos/server/internal/database"
	"github.com/Sambit-Mondal/karbos/server/internal/handlers"
	"github.com/Sambit-Mondal/karbos/server/internal/metrics"
	"github.com/Sambit-Mondal/karbos/server/internal/queue"
	"github.com/Sambit-Mondal/karbos/server/internal/scheduler"
	"github.com/Sambit-Mondal/karbos/server/internal/worker"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Println("🚀 Starting Karbos Server...")
	log.Printf("Environment: %s", cfg.Server.Environment)

	// Initialize database
	db, err := database.NewDatabase(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis queue
	redisQueue, err := queue.NewRedisQueue(
		cfg.GetRedisAddr(),
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Queue.ImmediateQueueKey,
		cfg.Queue.DelayedSetKey,
	)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisQueue.Close()

	// Initialize repositories
	jobRepo := database.NewJobRepository(db)
	carbonCacheRepo := database.NewCarbonCacheRepository(db)

	// Initialize carbon service
	var carbonService carbon.CarbonService
	cacheTTL, _ := time.ParseDuration(cfg.Carbon.CacheTTL)
	if cacheTTL == 0 {
		cacheTTL = 1 * time.Hour
	}

	if cfg.Carbon.Provider == "watttime" && cfg.Carbon.APIUsername != "" {
		log.Println("✓ Using WattTime carbon service")
		wattTimeClient := carbon.NewWattTimeClient(
			cfg.Carbon.APIUsername,
			cfg.Carbon.APIPassword,
			cfg.Carbon.BaseURL,
		)
		// Wrap with circuit breaker
		carbonService = wrapWithCircuitBreaker(wattTimeClient, cfg)
	} else if cfg.Carbon.APIKey != "" {
		log.Println("✓ Using ElectricityMaps carbon service")
		emClient := carbon.NewElectricityMapsClient(
			cfg.Carbon.APIKey,
			cfg.Carbon.BaseURL,
		)
		// Wrap with circuit breaker
		carbonService = wrapWithCircuitBreaker(emClient, cfg)
	} else {
		log.Println("⚠ No carbon API configured, scheduling will use default behavior")
	}

	// Initialize carbon fetcher with cache
	var carbonFetcher *carbon.CarbonFetcher
	var carbonScheduler *scheduler.CarbonScheduler

	if carbonService != nil {
		cacheWrapper := carbon.NewDatabaseCacheWrapper(carbonCacheRepo)
		carbonFetcher = carbon.NewCarbonFetcher(carbonService, cacheWrapper, cacheTTL)
		carbonScheduler = scheduler.NewCarbonScheduler(carbonFetcher)
		log.Println("✓ Carbon-aware scheduling enabled")
	}

	// Initialize delayed job promoter
	promoterCheckInterval, _ := time.ParseDuration(cfg.Promoter.CheckInterval)
	if promoterCheckInterval == 0 {
		promoterCheckInterval = 10 * time.Second
	}
	promoterService := worker.NewPromoterService(redisQueue, promoterCheckInterval)

	// Start promoter service
	ctx := context.Background()
	if err := promoterService.Start(ctx); err != nil {
		log.Fatalf("Failed to start promoter service: %v", err)
	}
	defer promoterService.Stop()

	// Initialize Prometheus metrics (if enabled)
	var metricsCollector *metrics.MetricsCollector
	if cfg.Metrics.Enabled {
		metricsCollector = metrics.NewMetricsCollector(redisQueue, nil, db.DB) // workerPool will be nil (API server doesn't run workers)
		// Start background metrics updater (every 10 seconds)
		metricsCollector.StartBackgroundUpdater(ctx, 10*time.Second)
		log.Printf("✓ Prometheus metrics enabled on port %s", cfg.Metrics.Port)
	}

	// Initialize handlers
	jobHandler := handlers.NewJobHandler(jobRepo, redisQueue, carbonScheduler)
	healthHandler := handlers.NewHealthHandler(db, redisQueue)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:               "Karbos API Gateway v1.0",
		ServerHeader:          "Karbos",
		DisableStartupMessage: false,
		ErrorHandler:          customErrorHandler,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           120 * time.Second,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(requestid.New())

	// CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Logger middleware (only in development)
	if cfg.IsDevelopment() {
		app.Use(logger.New(logger.Config{
			Format:     "[${time}] ${status} - ${method} ${path} (${latency})\n",
			TimeFormat: "15:04:05",
			TimeZone:   "Local",
		}))
	}

	// Routes
	setupRoutes(app, jobHandler, healthHandler, metricsCollector, cfg)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("\n🛑 Shutting down server gracefully...")

		// Give outstanding requests 10 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}

		log.Println("✓ Server stopped")
	}()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("✓ Server listening on http://localhost%s", addr)
	log.Println("✓ Phase 5: Reliability & Monitoring Complete!")
	log.Println("✓ All 5 Phases Operational - Production-Ready Carbon-Aware Job Scheduling System!")
	log.Println("\n📋 Available Endpoints:")
	log.Println("  POST   /api/submit          - Submit a new job (with carbon-aware scheduling)")
	log.Println("  GET    /api/jobs/:id        - Get job details")
	log.Println("  GET    /api/users/:id/jobs  - Get user's jobs")
	log.Println("  GET    /health              - Health check")
	log.Println("  GET    /ready               - Readiness check")
	if cfg.Metrics.Enabled {
		log.Printf("  GET    /metrics             - Prometheus metrics (port %s)\n", cfg.Metrics.Port)
	}

	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// wrapWithCircuitBreaker wraps a carbon service with circuit breaker protection
func wrapWithCircuitBreaker(service carbon.CarbonService, cfg *config.Config) carbon.CarbonService {
	timeout, _ := time.ParseDuration(cfg.CircuitBreaker.Timeout)
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	resetTimeout, _ := time.ParseDuration(cfg.CircuitBreaker.ResetTimeout)
	if resetTimeout == 0 {
		resetTimeout = 10 * time.Second
	}

	var staticFallback float64
	if _, err := fmt.Sscanf(cfg.CircuitBreaker.StaticFallback, "%f", &staticFallback); err != nil {
		staticFallback = 400.0 // Default global average
	}

	cbConfig := carbon.CircuitBreakerConfig{
		MaxFailures:    cfg.CircuitBreaker.MaxFailures,
		Timeout:        timeout,
		ResetTimeout:   resetTimeout,
		StaticFallback: staticFallback,
	}

	circuitBreaker := carbon.NewCircuitBreaker(service, cbConfig)
	log.Printf("✓ Circuit breaker enabled (threshold: %d failures, fallback: %.1f gCO2eq/kWh)",
		cbConfig.MaxFailures, cbConfig.StaticFallback)

	return circuitBreaker
}

// setupRoutes configures all API routes
func setupRoutes(app *fiber.App, jobHandler *handlers.JobHandler, healthHandler *handlers.HealthHandler, metricsCollector *metrics.MetricsCollector, cfg *config.Config) {
	// Health checks
	app.Get("/health", healthHandler.HealthCheck)
	app.Get("/ready", healthHandler.ReadyCheck)

	// Metrics endpoint (if enabled)
	if cfg.Metrics.Enabled && metricsCollector != nil {
		app.Get("/metrics", func(c *fiber.Ctx) error {
			// Update metrics before serving
			ctx := context.Background()
			if err := metricsCollector.UpdateMetrics(ctx); err != nil {
				log.Printf("Warning: Failed to update metrics: %v", err)
			}

			// Get metrics as Prometheus text format
			// We'll use the promhttp handler by creating an adapter
			return c.SendString(metricsCollector.GetPrometheusText())
		})
	}

	// API v1 routes
	api := app.Group("/api")

	// Job routes
	api.Post("/submit", jobHandler.SubmitJob)
	api.Get("/jobs/:id", jobHandler.GetJob)
	api.Get("/users/:userId/jobs", jobHandler.GetUserJobs)

	// Root endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "Karbos API Gateway",
			"version": "1.0.0",
			"status":  "operational",
			"phase":   "5 - Reliability & Monitoring (Production Ready)",
			"features": []string{
				"Phase 1: Infrastructure Skeleton",
				"Phase 2: Worker Node (Docker Execution)",
				"Phase 3: Intelligence Layer (Carbon-Aware Scheduling)",
				"Phase 4: The Dispatcher (Time-Based Job Promotion)",
				"Phase 5: Reliability & Monitoring (Circuit Breaker, Metrics, Graceful Shutdown)",
			},
		})
	})

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "not_found",
			"message": "Endpoint not found",
			"code":    404,
		})
	})
}

// customErrorHandler handles errors globally
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	log.Printf("Error: %v", err)

	return c.Status(code).JSON(fiber.Map{
		"error":   "server_error",
		"message": err.Error(),
		"code":    code,
	})
}
