package carbon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CarbonService interface for multiple provider support
type CarbonService interface {
	GetCarbonIntensity(ctx context.Context, region string, timestamp time.Time) (*CarbonIntensity, error)
	GetCarbonForecast(ctx context.Context, region string, startTime, endTime time.Time) ([]CarbonIntensity, error)
}

// CarbonIntensity represents carbon intensity data
type CarbonIntensity struct {
	Region          string    `json:"region"`
	Timestamp       time.Time `json:"timestamp"`
	Intensity       float64   `json:"intensity"`        // gCO2eq/kWh
	Unit            string    `json:"unit"`             // "gCO2eq/kWh"
	FossilFuel      float64   `json:"fossil_fuel"`      // Percentage
	RenewableEnergy float64   `json:"renewable_energy"` // Percentage
}

// ElectricityMapsClient implements CarbonService for ElectricityMaps API
type ElectricityMapsClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewElectricityMapsClient creates a new ElectricityMaps API client
func NewElectricityMapsClient(apiKey string, baseURL string) *ElectricityMapsClient {
	if baseURL == "" {
		baseURL = "https://api.electricitymap.org/v3"
	}
	return &ElectricityMapsClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ElectricityMapsResponse structure for current carbon intensity
type ElectricityMapsResponse struct {
	Zone                 string  `json:"zone"`
	CarbonIntensity      float64 `json:"carbonIntensity"`
	Datetime             string  `json:"datetime"`
	FossilFreePercentage float64 `json:"fossilFreePercentage"`
}

// ElectricityMapsForecastResponse structure for forecast data
type ElectricityMapsForecastResponse struct {
	Zone     string                         `json:"zone"`
	Forecast []ElectricityMapsForecastPoint `json:"forecast"`
}

// ElectricityMapsForecastPoint represents a single forecast data point
type ElectricityMapsForecastPoint struct {
	CarbonIntensity float64 `json:"carbonIntensity"`
	Datetime        string  `json:"datetime"`
}

// GetCarbonIntensity retrieves current carbon intensity for a region
func (c *ElectricityMapsClient) GetCarbonIntensity(ctx context.Context, region string, timestamp time.Time) (*CarbonIntensity, error) {
	// ElectricityMaps API endpoint: /carbon-intensity/latest?zone={zone}
	url := fmt.Sprintf("%s/carbon-intensity/latest?zone=%s", c.baseURL, region)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("auth-token", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp ElectricityMapsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Parse datetime
	parsedTime, err := time.Parse(time.RFC3339, apiResp.Datetime)
	if err != nil {
		parsedTime = time.Now()
	}

	return &CarbonIntensity{
		Region:          apiResp.Zone,
		Timestamp:       parsedTime,
		Intensity:       apiResp.CarbonIntensity,
		Unit:            "gCO2eq/kWh",
		RenewableEnergy: apiResp.FossilFreePercentage,
		FossilFuel:      100 - apiResp.FossilFreePercentage,
	}, nil
}

// GetCarbonForecast retrieves carbon intensity forecast for a region over a time range
func (c *ElectricityMapsClient) GetCarbonForecast(ctx context.Context, region string, startTime, endTime time.Time) ([]CarbonIntensity, error) {
	// ElectricityMaps API endpoint: /carbon-intensity/forecast?zone={zone}
	url := fmt.Sprintf("%s/carbon-intensity/forecast?zone=%s", c.baseURL, region)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("auth-token", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp ElectricityMapsForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert forecast points to CarbonIntensity objects
	var result []CarbonIntensity
	for _, point := range apiResp.Forecast {
		parsedTime, err := time.Parse(time.RFC3339, point.Datetime)
		if err != nil {
			continue // Skip invalid timestamps
		}

		// Filter by time range
		if parsedTime.Before(startTime) || parsedTime.After(endTime) {
			continue
		}

		result = append(result, CarbonIntensity{
			Region:    apiResp.Zone,
			Timestamp: parsedTime,
			Intensity: point.CarbonIntensity,
			Unit:      "gCO2eq/kWh",
		})
	}

	return result, nil
}

// WattTimeClient implements CarbonService for WattTime API (alternative provider)
type WattTimeClient struct {
	username    string
	password    string
	baseURL     string
	httpClient  *http.Client
	token       string
	tokenExpiry time.Time
}

// NewWattTimeClient creates a new WattTime API client
func NewWattTimeClient(username, password, baseURL string) *WattTimeClient {
	if baseURL == "" {
		baseURL = "https://api2.watttime.org/v2"
	}
	return &WattTimeClient{
		username: username,
		password: password,
		baseURL:  baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// authenticate retrieves an access token from WattTime API
func (w *WattTimeClient) authenticate(ctx context.Context) error {
	if w.token != "" && time.Now().Before(w.tokenExpiry) {
		return nil // Token still valid
	}

	url := fmt.Sprintf("%s/login", w.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.SetBasicAuth(w.username, w.password)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	w.token = authResp.Token
	w.tokenExpiry = time.Now().Add(30 * time.Minute)
	return nil
}

// GetCarbonIntensity retrieves current carbon intensity from WattTime
func (w *WattTimeClient) GetCarbonIntensity(ctx context.Context, region string, timestamp time.Time) (*CarbonIntensity, error) {
	if err := w.authenticate(ctx); err != nil {
		return nil, err
	}

	// WattTime uses "ba" (balancing authority) instead of zone
	url := fmt.Sprintf("%s/index?ba=%s", w.baseURL, region)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+w.token)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		BA      string  `json:"ba"`
		Percent float64 `json:"percent"` // 0-100 scale
		Point   string  `json:"point_time"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	parsedTime, err := time.Parse(time.RFC3339, apiResp.Point)
	if err != nil {
		parsedTime = time.Now()
	}

	// WattTime returns relative index (0-100), convert to approximate gCO2eq/kWh
	// Assuming max intensity ~800 gCO2eq/kWh for scaling
	intensity := (apiResp.Percent / 100.0) * 800.0

	return &CarbonIntensity{
		Region:    apiResp.BA,
		Timestamp: parsedTime,
		Intensity: intensity,
		Unit:      "gCO2eq/kWh",
	}, nil
}

// GetCarbonForecast retrieves forecast data from WattTime
func (w *WattTimeClient) GetCarbonForecast(ctx context.Context, region string, startTime, endTime time.Time) ([]CarbonIntensity, error) {
	if err := w.authenticate(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/forecast?ba=%s", w.baseURL, region)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+w.token)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp []struct {
		BA      string  `json:"ba"`
		Percent float64 `json:"percent"`
		Point   string  `json:"point_time"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var result []CarbonIntensity
	for _, point := range apiResp {
		parsedTime, err := time.Parse(time.RFC3339, point.Point)
		if err != nil {
			continue
		}

		if parsedTime.Before(startTime) || parsedTime.After(endTime) {
			continue
		}

		intensity := (point.Percent / 100.0) * 800.0

		result = append(result, CarbonIntensity{
			Region:    point.BA,
			Timestamp: parsedTime,
			Intensity: intensity,
			Unit:      "gCO2eq/kWh",
		})
	}

	return result, nil
}
