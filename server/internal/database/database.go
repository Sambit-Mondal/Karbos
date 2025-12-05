package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// DB holds the database connection
type DB struct {
	*sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✓ Successfully connected to PostgreSQL database")

	return &DB{db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	log.Println("Closing database connection...")
	return db.DB.Close()
}

// HealthCheck performs a database health check
func (db *DB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}
