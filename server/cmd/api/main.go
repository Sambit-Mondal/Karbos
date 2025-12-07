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
		carbonService = carbon.NewWattTimeClient(
			cfg.Carbon.APIUsername,
			cfg.Carbon.APIPassword,
			cfg.Carbon.BaseURL,
		)
	} else if cfg.Carbon.APIKey != "" {
		log.Println("✓ Using ElectricityMaps carbon service")
		carbonService = carbon.NewElectricityMapsClient(
			cfg.Carbon.APIKey,
			cfg.Carbon.BaseURL,
		)
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
	setupRoutes(app, jobHandler, healthHandler)

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
	log.Println("✓ Phase 3 Intelligence Layer Complete!")
	log.Println("\n📋 Available Endpoints:")
	log.Println("  POST   /api/submit          - Submit a new job (with carbon-aware scheduling)")
	log.Println("  GET    /api/jobs/:id        - Get job details")
	log.Println("  GET    /api/users/:id/jobs  - Get user's jobs")
	log.Println("  GET    /health              - Health check")
	log.Println("  GET    /ready               - Readiness check")

	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupRoutes configures all API routes
func setupRoutes(app *fiber.App, jobHandler *handlers.JobHandler, healthHandler *handlers.HealthHandler) {
	// Health checks
	app.Get("/health", healthHandler.HealthCheck)
	app.Get("/ready", healthHandler.ReadyCheck)

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
			"phase":   "3 - Intelligence Layer (Carbon-Aware Scheduling)",
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
