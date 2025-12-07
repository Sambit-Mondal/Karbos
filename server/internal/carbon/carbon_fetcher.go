package carbon

import (
	"context"
	"fmt"
	"time"
)

// CarbonCacheEntry represents cached carbon data
type CarbonCacheEntry struct {
	Region    string
	Timestamp time.Time
	Intensity float64
	Unit      string
	FetchedAt time.Time
	ExpiresAt time.Time
}

// CacheRepository interface for carbon cache operations
type CacheRepository interface {
	GetCarbonIntensity(ctx context.Context, region string, timestamp time.Time) (*CarbonCacheEntry, error)
	GetCarbonForecast(ctx context.Context, region string, startTime, endTime time.Time) ([]CarbonCacheEntry, error)
	SaveCarbonIntensity(ctx context.Context, data *CarbonIntensity, ttl time.Duration) error
	BulkSaveCarbonIntensities(ctx context.Context, data []CarbonIntensity, ttl time.Duration) error
	IsCacheFresh(entry *CarbonCacheEntry, maxAge time.Duration) bool
}

// CarbonFetcher provides cache-first carbon intensity fetching
type CarbonFetcher struct {
	service     CarbonService
	cache       CacheRepository
	cacheTTL    time.Duration
	maxCacheAge time.Duration
}

// NewCarbonFetcher creates a new carbon intensity fetcher with caching
func NewCarbonFetcher(service CarbonService, cache CacheRepository, cacheTTL time.Duration) *CarbonFetcher {
	if cacheTTL == 0 {
		cacheTTL = 1 * time.Hour // Default 1 hour TTL
	}
	return &CarbonFetcher{
		service:     service,
		cache:       cache,
		cacheTTL:    cacheTTL,
		maxCacheAge: cacheTTL,
	}
}

// GetCarbonIntensity retrieves carbon intensity with cache-first logic
// 1. Check cache for data
// 2. If cache hit and fresh (< 1 hour), return cached data
// 3. If cache miss or stale, fetch from API
// 4. Save API response to cache
func (f *CarbonFetcher) GetCarbonIntensity(ctx context.Context, region string, timestamp time.Time) (*CarbonIntensity, error) {
	// Step 1: Try cache first
	cachedEntry, err := f.cache.GetCarbonIntensity(ctx, region, timestamp)
	if err != nil {
		// Log error but continue to API fallback
		fmt.Printf("Cache error (continuing to API): %v\n", err)
	}

	// Step 2: Check cache freshness
	if cachedEntry != nil && f.cache.IsCacheFresh(cachedEntry, f.maxCacheAge) {
		// Cache hit with fresh data
		return &CarbonIntensity{
			Region:    cachedEntry.Region,
			Timestamp: cachedEntry.Timestamp,
			Intensity: cachedEntry.Intensity,
			Unit:      cachedEntry.Unit,
		}, nil
	}

	// Step 3: Cache miss or stale - fetch from API
	apiData, err := f.service.GetCarbonIntensity(ctx, region, timestamp)
	if err != nil {
		// If API fails but we have stale cache data, use it as fallback
		if cachedEntry != nil {
			fmt.Printf("API error (using stale cache): %v\n", err)
			return &CarbonIntensity{
				Region:    cachedEntry.Region,
				Timestamp: cachedEntry.Timestamp,
				Intensity: cachedEntry.Intensity,
				Unit:      cachedEntry.Unit,
			}, nil
		}
		return nil, fmt.Errorf("failed to fetch carbon intensity from API: %w", err)
	}

	// Step 4: Save fresh data to cache
	if err := f.cache.SaveCarbonIntensity(ctx, apiData, f.cacheTTL); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to save to cache: %v\n", err)
	}

	return apiData, nil
}

// GetCarbonForecast retrieves carbon intensity forecast with cache-first logic
func (f *CarbonFetcher) GetCarbonForecast(ctx context.Context, region string, startTime, endTime time.Time) ([]CarbonIntensity, error) {
	// Step 1: Try cache first
	cachedEntries, err := f.cache.GetCarbonForecast(ctx, region, startTime, endTime)
	if err != nil {
		fmt.Printf("Cache error (continuing to API): %v\n", err)
	}

	// Step 2: Check if cache has sufficient coverage
	// We need at least 80% coverage of the requested time range
	requiredDataPoints := int(endTime.Sub(startTime).Hours())
	if len(cachedEntries) >= int(float64(requiredDataPoints)*0.8) {
		// Check if all cached entries are fresh
		allFresh := true
		for _, entry := range cachedEntries {
			if !f.cache.IsCacheFresh(&CarbonCacheEntry{
				FetchedAt: entry.FetchedAt,
				ExpiresAt: entry.ExpiresAt,
			}, f.maxCacheAge) {
				allFresh = false
				break
			}
		}

		if allFresh {
			// Convert cache entries to CarbonIntensity
			var result []CarbonIntensity
			for _, entry := range cachedEntries {
				result = append(result, CarbonIntensity{
					Region:    entry.Region,
					Timestamp: entry.Timestamp,
					Intensity: entry.Intensity,
					Unit:      entry.Unit,
				})
			}
			return result, nil
		}
	}

	// Step 3: Cache miss or insufficient coverage - fetch from API
	apiData, err := f.service.GetCarbonForecast(ctx, region, startTime, endTime)
	if err != nil {
		// If API fails but we have some cache data, use it as fallback
		if len(cachedEntries) > 0 {
			fmt.Printf("API error (using partial cache): %v\n", err)
			var result []CarbonIntensity
			for _, entry := range cachedEntries {
				result = append(result, CarbonIntensity{
					Region:    entry.Region,
					Timestamp: entry.Timestamp,
					Intensity: entry.Intensity,
					Unit:      entry.Unit,
				})
			}
			return result, nil
		}
		return nil, fmt.Errorf("failed to fetch carbon forecast from API: %w", err)
	}

	// Step 4: Bulk save fresh data to cache
	if err := f.cache.BulkSaveCarbonIntensities(ctx, apiData, f.cacheTTL); err != nil {
		fmt.Printf("Failed to save forecast to cache: %v\n", err)
	}

	return apiData, nil
}

// GetCurrentCarbonIntensity is a convenience method to get current carbon intensity
func (f *CarbonFetcher) GetCurrentCarbonIntensity(ctx context.Context, region string) (*CarbonIntensity, error) {
	return f.GetCarbonIntensity(ctx, region, time.Now())
}

// GetForecastForWindow retrieves carbon forecast for a specific window size
func (f *CarbonFetcher) GetForecastForWindow(ctx context.Context, region string, windowHours int) ([]CarbonIntensity, error) {
	now := time.Now()
	endTime := now.Add(time.Duration(windowHours) * time.Hour)
	return f.GetCarbonForecast(ctx, region, now, endTime)
}
