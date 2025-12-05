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
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port        string
	Environment string
	RateLimit   string
	Timeout     string
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

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// GetRedisAddr returns the full Redis address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}
