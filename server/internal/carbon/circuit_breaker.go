package carbon

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	// StateClosed - circuit is closed, requests pass through
	StateClosed CircuitState = iota
	// StateOpen - circuit is open, requests fail fast
	StateOpen
	// StateHalfOpen - circuit is testing if service recovered
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig holds configuration for the circuit breaker
type CircuitBreakerConfig struct {
	MaxFailures    int           // Number of failures before opening circuit
	Timeout        time.Duration // How long to wait before trying again (open -> half-open)
	ResetTimeout   time.Duration // How long to stay in half-open before closing
	StaticFallback float64       // Static carbon intensity value when circuit is open (gCO2eq/kWh)
	StaticRegion   string        // Default region for static fallback
}

// CircuitBreaker wraps a CarbonService with circuit breaker pattern
type CircuitBreaker struct {
	service CarbonService
	config  CircuitBreakerConfig

	mu            sync.RWMutex
	state         CircuitState
	failures      int
	lastFailTime  time.Time
	lastStateTime time.Time
	successCount  int // Track successes in half-open state
}

// NewCircuitBreaker creates a new circuit breaker for carbon service
func NewCircuitBreaker(service CarbonService, config CircuitBreakerConfig) *CircuitBreaker {
	if config.MaxFailures == 0 {
		config.MaxFailures = 5 // Default: 5 failures
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second // Default: 30 seconds
	}
	if config.ResetTimeout == 0 {
		config.ResetTimeout = 10 * time.Second // Default: 10 seconds
	}
	if config.StaticFallback == 0 {
		config.StaticFallback = 400.0 // Default: 400 gCO2eq/kWh (global average)
	}
	if config.StaticRegion == "" {
		config.StaticRegion = "GLOBAL-AVERAGE"
	}

	return &CircuitBreaker{
		service:       service,
		config:        config,
		state:         StateClosed,
		lastStateTime: time.Now(),
	}
}

// GetCarbonIntensity retrieves carbon intensity with circuit breaker protection
func (cb *CircuitBreaker) GetCarbonIntensity(ctx context.Context, region string, timestamp time.Time) (*CarbonIntensity, error) {
	// Check circuit state
	if !cb.canAttempt() {
		// Circuit is open - return static fallback
		return cb.fallbackIntensity(region, timestamp), nil
	}

	// Attempt to call underlying service
	result, err := cb.service.GetCarbonIntensity(ctx, region, timestamp)

	if err != nil {
		cb.recordFailure(err)
		// Return fallback on error
		return cb.fallbackIntensity(region, timestamp), nil
	}

	// Success - record it
	cb.recordSuccess()
	return result, nil
}

// GetCarbonForecast retrieves carbon forecast with circuit breaker protection
func (cb *CircuitBreaker) GetCarbonForecast(ctx context.Context, region string, startTime, endTime time.Time) ([]CarbonIntensity, error) {
	// Check circuit state
	if !cb.canAttempt() {
		// Circuit is open - return static fallback forecast
		return cb.fallbackForecast(region, startTime, endTime), nil
	}

	// Attempt to call underlying service
	result, err := cb.service.GetCarbonForecast(ctx, region, startTime, endTime)

	if err != nil {
		cb.recordFailure(err)
		// Return fallback on error
		return cb.fallbackForecast(region, startTime, endTime), nil
	}

	// Success - record it
	cb.recordSuccess()
	return result, nil
}

// canAttempt checks if a request can be attempted based on circuit state
func (cb *CircuitBreaker) canAttempt() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		// Circuit closed - allow request
		return true

	case StateOpen:
		// Check if timeout has elapsed
		if now.Sub(cb.lastStateTime) >= cb.config.Timeout {
			// Transition to half-open
			cb.state = StateHalfOpen
			cb.lastStateTime = now
			cb.successCount = 0
			fmt.Printf("🔧 Circuit breaker transitioning to HALF_OPEN (will test service recovery)\n")
			return true
		}
		// Still in timeout - reject request
		return false

	case StateHalfOpen:
		// In half-open state - allow single request to test
		return true

	default:
		return false
	}
}

// recordFailure records a failed request
func (cb *CircuitBreaker) recordFailure(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	cb.failures++
	cb.lastFailTime = now

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			// Open the circuit
			cb.state = StateOpen
			cb.lastStateTime = now
			fmt.Printf("🚨 Circuit breaker OPENED after %d failures (last error: %v)\n", cb.failures, err)
			fmt.Printf("⚠️  Using static fallback: %.2f gCO2eq/kWh for next %v\n", cb.config.StaticFallback, cb.config.Timeout)
		} else {
			fmt.Printf("⚠️  Carbon API failure %d/%d: %v\n", cb.failures, cb.config.MaxFailures, err)
		}

	case StateHalfOpen:
		// Failed in half-open - back to open
		cb.state = StateOpen
		cb.lastStateTime = now
		cb.failures = cb.config.MaxFailures // Reset to max
		fmt.Printf("🚨 Circuit breaker back to OPEN (service still failing)\n")
	}
}

// recordSuccess records a successful request
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		// Already closed - reset failure count
		if cb.failures > 0 {
			fmt.Printf("✓ Carbon API recovered (resetting failure count: %d → 0)\n", cb.failures)
			cb.failures = 0
		}

	case StateHalfOpen:
		cb.successCount++
		// After successful request in half-open, close the circuit
		cb.state = StateClosed
		cb.failures = 0
		cb.lastStateTime = time.Now()
		fmt.Printf("✓ Circuit breaker CLOSED (service recovered after half-open test)\n")
	}
}

// fallbackIntensity returns a static fallback carbon intensity
func (cb *CircuitBreaker) fallbackIntensity(region string, timestamp time.Time) *CarbonIntensity {
	return &CarbonIntensity{
		Region:    region,
		Timestamp: timestamp,
		Intensity: cb.config.StaticFallback,
		Unit:      "gCO2eq/kWh",
	}
}

// fallbackForecast returns a static fallback forecast
func (cb *CircuitBreaker) fallbackForecast(region string, startTime, endTime time.Time) []CarbonIntensity {
	var forecast []CarbonIntensity

	// Generate hourly data points with static intensity
	current := startTime
	for current.Before(endTime) {
		forecast = append(forecast, CarbonIntensity{
			Region:    region,
			Timestamp: current,
			Intensity: cb.config.StaticFallback,
			Unit:      "gCO2eq/kWh",
		})
		current = current.Add(1 * time.Hour)
	}

	return forecast
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetFailures returns the current failure count
func (cb *CircuitBreaker) GetFailures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":                cb.state.String(),
		"failures":             cb.failures,
		"max_failures":         cb.config.MaxFailures,
		"last_fail_time":       cb.lastFailTime,
		"last_state_change":    cb.lastStateTime,
		"timeout":              cb.config.Timeout.String(),
		"static_fallback":      cb.config.StaticFallback,
		"success_count":        cb.successCount,
		"time_since_last_fail": time.Since(cb.lastFailTime).String(),
	}
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successCount = 0
	cb.lastStateTime = time.Now()
	fmt.Println("✓ Circuit breaker manually reset to CLOSED state")
}
