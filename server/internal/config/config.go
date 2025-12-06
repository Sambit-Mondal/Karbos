package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Queue    QueueConfig
	Worker   WorkerConfig
	Docker   DockerConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port        string
	Environment string
	RateLimit   string
	Timeout     string
}

// WorkerConfig holds worker pool configuration
type WorkerConfig struct {
	PoolSize     int
	PollInterval string
	JobTimeout   string
	MaxRetries   int
}

// DockerConfig holds Docker daemon configuration
type DockerConfig struct {
	Host        string
	MemoryLimit int64
	CPUQuota    int64
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	URL string
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// QueueConfig holds queue-specific configuration
type QueueConfig struct {
	ImmediateQueueKey string
	DelayedSetKey     string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists (optional in production)
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	config := &Config{
		Server: ServerConfig{
			Port:        getEnv("PORT", "8080"),
			Environment: getEnv("ENV", "development"),
			RateLimit:   getEnv("API_RATE_LIMIT", "100"),
			Timeout:     getEnv("API_TIMEOUT", "30s"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", ""),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		Queue: QueueConfig{
			ImmediateQueueKey: getEnv("IMMEDIATE_QUEUE_KEY", "karbos:queue:immediate"),
			DelayedSetKey:     getEnv("DELAYED_SET_KEY", "karbos:queue:delayed"),
		},
		Worker: WorkerConfig{
			PoolSize:     getEnvAsInt("WORKER_POOL_SIZE", 5),
			PollInterval: getEnv("WORKER_POLL_INTERVAL", "2s"),
			JobTimeout:   getEnv("WORKER_JOB_TIMEOUT", "10m"),
			MaxRetries:   getEnvAsInt("WORKER_MAX_RETRIES", 3),
		},
		Docker: DockerConfig{
			Host:        getEnv("DOCKER_HOST", ""),
			MemoryLimit: getEnvAsInt64("DOCKER_MEMORY_LIMIT", 536870912), // 512MB
			CPUQuota:    getEnvAsInt64("DOCKER_CPU_QUOTA", 50000),        // 50% of one CPU
		},
	}

	// Validate required configuration
	if config.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return config, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt retrieves an environment variable as int or returns default
func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var intValue int
	if _, err := fmt.Sscanf(value, "%d", &intValue); err != nil {
		log.Printf("Warning: Invalid integer value for %s, using default: %d", key, defaultValue)
		return defaultValue
	}
	return intValue
}

// getEnvAsInt64 retrieves an environment variable as int64 or returns default
func getEnvAsInt64(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var intValue int64
	if _, err := fmt.Sscanf(value, "%d", &intValue); err != nil {
		log.Printf("Warning: Invalid int64 value for %s, using default: %d", key, defaultValue)
		return defaultValue
	}
	return intValue
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// GetRedisAddr returns the full Redis address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}
