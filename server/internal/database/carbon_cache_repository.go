package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CarbonCacheRepository handles carbon cache database operations
type CarbonCacheRepository struct {
	db *DB
}

// NewCarbonCacheRepository creates a new carbon cache repository
func NewCarbonCacheRepository(db *DB) *CarbonCacheRepository {
	return &CarbonCacheRepository{db: db}
}

// CarbonCacheEntry represents a cached carbon intensity record
type CarbonCacheEntry struct {
	ID        uuid.UUID `json:"id"`
	Region    string    `json:"region"`
	Timestamp time.Time `json:"timestamp"`
	Intensity float64   `json:"intensity"`
	Unit      string    `json:"unit"`
	FetchedAt time.Time `json:"fetched_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CarbonIntensity is a local type for saving data (avoids circular import)
type CarbonIntensity struct {
	Region    string
	Timestamp time.Time
	Intensity float64
	Unit      string
}

// SaveCarbonIntensity saves carbon intensity data to cache
func (r *CarbonCacheRepository) SaveCarbonIntensity(ctx context.Context, region string, timestamp time.Time, intensity float64, unit string, ttl time.Duration) error {
	query := `
		INSERT INTO carbon_cache (id, region, timestamp, intensity, unit, fetched_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (region, timestamp) 
		DO UPDATE SET 
			intensity = EXCLUDED.intensity,
			unit = EXCLUDED.unit,
			fetched_at = EXCLUDED.fetched_at,
			expires_at = EXCLUDED.expires_at
	`

	id := uuid.New()
	now := time.Now()
	expiresAt := now.Add(ttl)

	_, err := r.db.ExecContext(ctx, query,
		id,
		region,
		timestamp,
		intensity,
		unit,
		now,
		expiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save carbon intensity to cache: %w", err)
	}

	return nil
}

// GetCarbonIntensity retrieves cached carbon intensity data
func (r *CarbonCacheRepository) GetCarbonIntensity(ctx context.Context, region string, timestamp time.Time) (*CarbonCacheEntry, error) {
	query := `
		SELECT id, region, timestamp, intensity, unit, fetched_at, expires_at
		FROM carbon_cache
		WHERE region = $1 
			AND timestamp >= $2 - INTERVAL '15 minutes'
			AND timestamp <= $2 + INTERVAL '15 minutes'
			AND expires_at > NOW()
		ORDER BY ABS(EXTRACT(EPOCH FROM (timestamp - $2)))
		LIMIT 1
	`

	var entry CarbonCacheEntry
	err := r.db.QueryRowContext(ctx, query, region, timestamp).Scan(
		&entry.ID,
		&entry.Region,
		&entry.Timestamp,
		&entry.Intensity,
		&entry.Unit,
		&entry.FetchedAt,
		&entry.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Cache miss
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get carbon intensity from cache: %w", err)
	}

	return &entry, nil
}

// GetCarbonForecast retrieves cached forecast data within a time range
func (r *CarbonCacheRepository) GetCarbonForecast(ctx context.Context, region string, startTime, endTime time.Time) ([]CarbonCacheEntry, error) {
	query := `
		SELECT id, region, timestamp, intensity, unit, fetched_at, expires_at
		FROM carbon_cache
		WHERE region = $1 
			AND timestamp BETWEEN $2 AND $3
			AND expires_at > NOW()
		ORDER BY timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, region, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get carbon forecast from cache: %w", err)
	}
	defer rows.Close()

	var entries []CarbonCacheEntry
	for rows.Next() {
		var entry CarbonCacheEntry
		err := rows.Scan(
			&entry.ID,
			&entry.Region,
			&entry.Timestamp,
			&entry.Intensity,
			&entry.Unit,
			&entry.FetchedAt,
			&entry.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan carbon cache entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// IsCacheFresh checks if cached data is still fresh based on TTL
func (r *CarbonCacheRepository) IsCacheFresh(entry *CarbonCacheEntry, maxAge time.Duration) bool {
	return time.Since(entry.FetchedAt) < maxAge && time.Now().Before(entry.ExpiresAt)
}

// DeleteExpiredEntries removes expired cache entries
func (r *CarbonCacheRepository) DeleteExpiredEntries(ctx context.Context) (int64, error) {
	query := `DELETE FROM carbon_cache WHERE expires_at <= NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired cache entries: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// GetCacheStats returns cache statistics
func (r *CarbonCacheRepository) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	var totalEntries, validEntries, expiredEntries int

	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE expires_at > NOW()) as valid,
			COUNT(*) FILTER (WHERE expires_at <= NOW()) as expired
		FROM carbon_cache
	`

	err := r.db.QueryRowContext(ctx, query).Scan(&totalEntries, &validEntries, &expiredEntries)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_entries":   totalEntries,
		"valid_entries":   validEntries,
		"expired_entries": expiredEntries,
	}

	return stats, nil
}

// BulkSaveCarbonIntensities saves multiple carbon intensity records
func (r *CarbonCacheRepository) BulkSaveCarbonIntensities(ctx context.Context, data []CarbonIntensity, ttl time.Duration) error {
	if len(data) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO carbon_cache (id, region, timestamp, intensity, unit, fetched_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (region, timestamp) 
		DO UPDATE SET 
			intensity = EXCLUDED.intensity,
			unit = EXCLUDED.unit,
			fetched_at = EXCLUDED.fetched_at,
			expires_at = EXCLUDED.expires_at
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	expiresAt := now.Add(ttl)

	for _, entry := range data {
		id := uuid.New()
		_, err := stmt.ExecContext(ctx,
			id,
			entry.Region,
			entry.Timestamp,
			entry.Intensity,
			entry.Unit,
			now,
			expiresAt,
		)
		if err != nil {
			return fmt.Errorf("failed to save entry: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
