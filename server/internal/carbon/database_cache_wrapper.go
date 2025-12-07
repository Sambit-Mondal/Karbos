package carbon

import (
	"context"
	"time"

	"github.com/Sambit-Mondal/karbos/server/internal/database"
)

// DatabaseCacheWrapper adapts database.CarbonCacheRepository to carbon.CacheRepository interface
type DatabaseCacheWrapper struct {
	repo *database.CarbonCacheRepository
}

// NewDatabaseCacheWrapper creates a new database cache wrapper
func NewDatabaseCacheWrapper(repo *database.CarbonCacheRepository) *DatabaseCacheWrapper {
	return &DatabaseCacheWrapper{repo: repo}
}

// GetCarbonIntensity retrieves cached carbon intensity data
func (w *DatabaseCacheWrapper) GetCarbonIntensity(ctx context.Context, region string, timestamp time.Time) (*CarbonCacheEntry, error) {
	dbEntry, err := w.repo.GetCarbonIntensity(ctx, region, timestamp)
	if err != nil {
		return nil, err
	}
	if dbEntry == nil {
		return nil, nil
	}

	return &CarbonCacheEntry{
		Region:    dbEntry.Region,
		Timestamp: dbEntry.Timestamp,
		Intensity: dbEntry.Intensity,
		Unit:      dbEntry.Unit,
		FetchedAt: dbEntry.FetchedAt,
		ExpiresAt: dbEntry.ExpiresAt,
	}, nil
}

// GetCarbonForecast retrieves cached forecast data
func (w *DatabaseCacheWrapper) GetCarbonForecast(ctx context.Context, region string, startTime, endTime time.Time) ([]CarbonCacheEntry, error) {
	dbEntries, err := w.repo.GetCarbonForecast(ctx, region, startTime, endTime)
	if err != nil {
		return nil, err
	}

	var entries []CarbonCacheEntry
	for _, dbEntry := range dbEntries {
		entries = append(entries, CarbonCacheEntry{
			Region:    dbEntry.Region,
			Timestamp: dbEntry.Timestamp,
			Intensity: dbEntry.Intensity,
			Unit:      dbEntry.Unit,
			FetchedAt: dbEntry.FetchedAt,
			ExpiresAt: dbEntry.ExpiresAt,
		})
	}

	return entries, nil
}

// SaveCarbonIntensity saves carbon intensity data to cache
func (w *DatabaseCacheWrapper) SaveCarbonIntensity(ctx context.Context, data *CarbonIntensity, ttl time.Duration) error {
	return w.repo.SaveCarbonIntensity(ctx, data.Region, data.Timestamp, data.Intensity, data.Unit, ttl)
}

// BulkSaveCarbonIntensities saves multiple carbon intensity records
func (w *DatabaseCacheWrapper) BulkSaveCarbonIntensities(ctx context.Context, data []CarbonIntensity, ttl time.Duration) error {
	dbData := make([]database.CarbonIntensity, len(data))
	for i, entry := range data {
		dbData[i] = database.CarbonIntensity{
			Region:    entry.Region,
			Timestamp: entry.Timestamp,
			Intensity: entry.Intensity,
			Unit:      entry.Unit,
		}
	}
	return w.repo.BulkSaveCarbonIntensities(ctx, dbData, ttl)
}

// IsCacheFresh checks if cached data is still fresh
func (w *DatabaseCacheWrapper) IsCacheFresh(entry *CarbonCacheEntry, maxAge time.Duration) bool {
	dbEntry := &database.CarbonCacheEntry{
		FetchedAt: entry.FetchedAt,
		ExpiresAt: entry.ExpiresAt,
	}
	return w.repo.IsCacheFresh(dbEntry, maxAge)
}
