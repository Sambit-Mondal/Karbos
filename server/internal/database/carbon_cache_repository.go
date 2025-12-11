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
	ID             uuid.UUID `json:"id"`
	Region         string    `json:"region"`
	Timestamp      time.Time `json:"timestamp"`
	IntensityValue float64   `json:"intensity_value"`
	ForecastWindow *int      `json:"forecast_window,omitempty"`
	Source         *string   `json:"source,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
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
		INSERT INTO carbon_cache (id, region, timestamp, intensity_value, source)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (region, timestamp, forecast_window) 
		DO UPDATE SET 
			intensity_value = EXCLUDED.intensity_value,
			source = EXCLUDED.source
	`

	id := uuid.New()
	source := "api"

	_, err := r.db.ExecContext(ctx, query,
		id,
		region,
		timestamp,
		intensity,
		&source,
	)

	if err != nil {
		return fmt.Errorf("failed to save carbon intensity to cache: %w", err)
	}

	return nil
}

// GetCarbonIntensity retrieves cached carbon intensity data
func (r *CarbonCacheRepository) GetCarbonIntensity(ctx context.Context, region string, timestamp time.Time) (*CarbonCacheEntry, error) {
	query := `
		SELECT id, region, timestamp, intensity_value, forecast_window, source, created_at
		FROM carbon_cache
		WHERE region = $1 
			AND timestamp >= $2 - INTERVAL '15 minutes'
			AND timestamp <= $2 + INTERVAL '15 minutes'
		ORDER BY ABS(EXTRACT(EPOCH FROM (timestamp - $2)))
		LIMIT 1
	`

	var entry CarbonCacheEntry
	err := r.db.QueryRowContext(ctx, query, region, timestamp).Scan(
		&entry.ID,
		&entry.Region,
		&entry.Timestamp,
		&entry.IntensityValue,
		&entry.ForecastWindow,
		&entry.Source,
		&entry.CreatedAt,
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
		SELECT id, region, timestamp, intensity_value, forecast_window, source, created_at
		FROM carbon_cache
		WHERE region = $1 
			AND timestamp BETWEEN $2 AND $3
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
			&entry.IntensityValue,
			&entry.ForecastWindow,
			&entry.Source,
			&entry.CreatedAt,
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
	return time.Since(entry.CreatedAt) < maxAge
}

// DeleteExpiredEntries removes expired cache entries older than maxAge
func (r *CarbonCacheRepository) DeleteExpiredEntries(ctx context.Context, maxAge time.Duration) (int64, error) {
	query := `DELETE FROM carbon_cache WHERE created_at <= NOW() - $1::INTERVAL`

	result, err := r.db.ExecContext(ctx, query, maxAge)
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
			COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '24 hours') as valid,
			COUNT(*) FILTER (WHERE created_at <= NOW() - INTERVAL '24 hours') as expired
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
		INSERT INTO carbon_cache (id, region, timestamp, intensity_value, source)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (region, timestamp, forecast_window) 
		DO UPDATE SET 
			intensity_value = EXCLUDED.intensity_value,
			source = EXCLUDED.source
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	source := "api"

	for _, entry := range data {
		id := uuid.New()
		_, err := stmt.ExecContext(ctx,
			id,
			entry.Region,
			entry.Timestamp,
			entry.Intensity,
			&source,
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

// GetRecentEntries retrieves all carbon cache entries from the last N duration
func (r *CarbonCacheRepository) GetRecentEntries(ctx context.Context, duration time.Duration) ([]CarbonCacheEntry, error) {
	query := `
		SELECT id, region, timestamp, intensity_value, forecast_window, source, created_at
		FROM carbon_cache
		WHERE timestamp >= NOW() - $1::interval
		ORDER BY timestamp DESC
		LIMIT 1000
	`

	rows, err := r.db.QueryContext(ctx, query, duration.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get recent cache entries: %w", err)
	}
	defer rows.Close()

	var entries []CarbonCacheEntry
	for rows.Next() {
		var entry CarbonCacheEntry
		err := rows.Scan(
			&entry.ID,
			&entry.Region,
			&entry.Timestamp,
			&entry.IntensityValue,
			&entry.ForecastWindow,
			&entry.Source,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan carbon cache entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetCarbonIntensityRange retrieves carbon intensity data for a specific region within a time range
func (r *CarbonCacheRepository) GetCarbonIntensityRange(ctx context.Context, region string, startTime, endTime time.Time) ([]CarbonCacheEntry, error) {
	query := `
		SELECT id, region, timestamp, intensity_value, forecast_window, source, created_at
		FROM carbon_cache
		WHERE region = $1 
			AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, region, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get carbon intensity range: %w", err)
	}
	defer rows.Close()

	var entries []CarbonCacheEntry
	for rows.Next() {
		var entry CarbonCacheEntry
		err := rows.Scan(
			&entry.ID,
			&entry.Region,
			&entry.Timestamp,
			&entry.IntensityValue,
			&entry.ForecastWindow,
			&entry.Source,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan carbon cache entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
