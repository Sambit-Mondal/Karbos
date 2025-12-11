package handlers

import (
	"context"
	"log"
	"time"

	"github.com/Sambit-Mondal/karbos/server/internal/database"
	"github.com/Sambit-Mondal/karbos/server/internal/models"
	"github.com/gofiber/fiber/v2"
)

// CarbonHandler handles carbon-related HTTP requests
type CarbonHandler struct {
	carbonRepo *database.CarbonCacheRepository
}

// NewCarbonHandler creates a new carbon handler
func NewCarbonHandler(carbonRepo *database.CarbonCacheRepository) *CarbonHandler {
	return &CarbonHandler{
		carbonRepo: carbonRepo,
	}
}

// CarbonForecastEntry represents a single forecast entry for the API
type CarbonForecastEntry struct {
	Region         string  `json:"region"`
	Timestamp      string  `json:"timestamp"`
	IntensityValue float64 `json:"intensity_value"`
	Unit           string  `json:"unit"`
}

// CarbonForecastResponse represents the carbon forecast API response
type CarbonForecastResponse struct {
	Region           string                `json:"region"`
	Forecasts        []CarbonForecastEntry `json:"forecasts"`
	CurrentIntensity *float64              `json:"current_intensity,omitempty"`
	OptimalTime      *string               `json:"optimal_time,omitempty"`
}

// GetCarbonForecast handles GET /api/carbon-forecast
func (h *CarbonHandler) GetCarbonForecast(c *fiber.Ctx) error {
	ctx := context.Background()

	// Get region from query params (default to all regions)
	region := c.Query("region", "")

	var cacheEntries []database.CarbonCacheEntry
	var err error

	if region != "" {
		// Get forecast for specific region
		// Fetch data for next 24 hours
		now := time.Now()
		endTime := now.Add(24 * time.Hour)

		cacheEntries, err = h.carbonRepo.GetCarbonIntensityRange(ctx, region, now, endTime)
		if err != nil {
			log.Printf("Failed to get carbon forecast for region %s: %v", region, err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Error:   "database_error",
				Message: "Failed to fetch carbon forecast",
				Code:    fiber.StatusInternalServerError,
			})
		}
	} else {
		// Get all recent carbon cache entries (last 24 hours across all regions)
		cacheEntries, err = h.carbonRepo.GetRecentEntries(ctx, 24*time.Hour)
		if err != nil {
			log.Printf("Failed to get recent carbon cache entries: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Error:   "database_error",
				Message: "Failed to fetch carbon cache data",
				Code:    fiber.StatusInternalServerError,
			})
		}
	}

	// Convert to API response format
	forecasts := make([]CarbonForecastEntry, len(cacheEntries))
	for i, entry := range cacheEntries {
		forecasts[i] = CarbonForecastEntry{
			Region:         entry.Region,
			Timestamp:      entry.Timestamp.Format(time.RFC3339),
			IntensityValue: entry.IntensityValue,
			Unit:           "gCO2/kWh",
		}
	}

	// Find current intensity (most recent entry) and optimal time (lowest intensity)
	var currentIntensity *float64
	var optimalTime *string
	var minIntensity *float64

	for i, entry := range forecasts {
		if i == 0 || (minIntensity != nil && entry.IntensityValue < *minIntensity) {
			minIntensity = &entry.IntensityValue
			optimalTime = &entry.Timestamp
		}

		// Set current intensity from most recent entry
		if currentIntensity == nil && region != "" {
			currentIntensity = &entry.IntensityValue
		}
	}

	response := CarbonForecastResponse{
		Region:           region,
		Forecasts:        forecasts,
		CurrentIntensity: currentIntensity,
		OptimalTime:      optimalTime,
	}

	return c.JSON(response)
}

// GetCarbonCache handles GET /api/carbon-cache
func (h *CarbonHandler) GetCarbonCache(c *fiber.Ctx) error {
	ctx := context.Background()

	// Get all recent entries (last 48 hours)
	cacheEntries, err := h.carbonRepo.GetRecentEntries(ctx, 48*time.Hour)
	if err != nil {
		log.Printf("Failed to get carbon cache entries: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch carbon cache",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Convert to API response format
	forecasts := make([]CarbonForecastEntry, len(cacheEntries))
	for i, entry := range cacheEntries {
		forecasts[i] = CarbonForecastEntry{
			Region:         entry.Region,
			Timestamp:      entry.Timestamp.Format(time.RFC3339),
			IntensityValue: entry.IntensityValue,
			Unit:           "gCO2/kWh",
		}
	}

	return c.JSON(forecasts)
}
