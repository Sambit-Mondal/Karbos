package scheduler

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Sambit-Mondal/karbos/server/internal/carbon"
)

// CarbonFetcher interface for retrieving carbon intensity data
type CarbonFetcher interface {
	GetCarbonForecast(ctx context.Context, region string, startTime, endTime time.Time) ([]carbon.CarbonIntensity, error)
	GetCurrentCarbonIntensity(ctx context.Context, region string) (*carbon.CarbonIntensity, error)
}

// ScheduleRequest represents a job scheduling request
type ScheduleRequest struct {
	Region       string        // Geographic region for carbon intensity
	Duration     time.Duration // Expected job execution duration
	Deadline     time.Time     // Latest time job must complete
	WindowSize   time.Duration // Time window to consider (default 24 hours)
	MinStartTime time.Time     // Earliest time job can start (default now)
}

// ScheduleResult contains the scheduling decision
type ScheduleResult struct {
	ScheduledTime      time.Time    // Optimal start time for job
	ExpectedIntensity  float64      // Expected carbon intensity at scheduled time
	Immediate          bool         // Whether to run immediately or schedule for later
	CarbonSavings      float64      // Estimated carbon savings vs immediate execution
	AlternativeWindows []TimeWindow // Other optimal windows
}

// TimeWindow represents a potential execution window
type TimeWindow struct {
	StartTime    time.Time
	EndTime      time.Time
	AvgIntensity float64
	CarbonCost   float64
}

// CarbonScheduler implements the sliding window scheduling algorithm
type CarbonScheduler struct {
	fetcher      CarbonFetcher
	slotDuration time.Duration // Duration of each time slot (default 1 hour)
	threshold    float64       // Carbon intensity threshold for immediate execution
}

// NewCarbonScheduler creates a new carbon-aware scheduler
func NewCarbonScheduler(fetcher CarbonFetcher) *CarbonScheduler {
	return &CarbonScheduler{
		fetcher:      fetcher,
		slotDuration: 1 * time.Hour,
		threshold:    400.0, // Default threshold: 400 gCO2eq/kWh
	}
}

// Schedule finds the optimal execution time for a job using sliding window algorithm
func (s *CarbonScheduler) Schedule(ctx context.Context, req *ScheduleRequest) (*ScheduleResult, error) {
	// Validate request
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	// Set defaults
	if req.WindowSize == 0 {
		req.WindowSize = 24 * time.Hour
	}
	if req.MinStartTime.IsZero() {
		req.MinStartTime = time.Now()
	}

	// Get carbon intensity forecast
	endTime := req.MinStartTime.Add(req.WindowSize)
	if endTime.After(req.Deadline) {
		endTime = req.Deadline
	}

	forecast, err := s.fetcher.GetCarbonForecast(ctx, req.Region, req.MinStartTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get carbon forecast: %w", err)
	}

	if len(forecast) == 0 {
		// No forecast data available - use current intensity
		current, err := s.fetcher.GetCurrentCarbonIntensity(ctx, req.Region)
		if err != nil {
			return nil, fmt.Errorf("failed to get current carbon intensity: %w", err)
		}
		return &ScheduleResult{
			ScheduledTime:     time.Now(),
			ExpectedIntensity: current.Intensity,
			Immediate:         true,
			CarbonSavings:     0,
		}, nil
	}

	// Run sliding window algorithm
	optimalWindow, alternativeWindows := s.findOptimalWindow(forecast, req.Duration, req.MinStartTime, req.Deadline)

	// Get current intensity for comparison
	currentIntensity := forecast[0].Intensity

	// Calculate carbon savings
	carbonSavings := currentIntensity - optimalWindow.AvgIntensity
	savingsPercent := (carbonSavings / currentIntensity) * 100

	// Decision: Immediate vs Scheduled
	immediate := false
	scheduledTime := optimalWindow.StartTime

	// Execute immediately if:
	// 1. Current time is already optimal
	// 2. Savings are negligible (< 10%)
	// 3. Current intensity is below threshold
	if time.Since(optimalWindow.StartTime) < 5*time.Minute ||
		savingsPercent < 10.0 ||
		currentIntensity < s.threshold {
		immediate = true
		scheduledTime = time.Now()
	}

	return &ScheduleResult{
		ScheduledTime:      scheduledTime,
		ExpectedIntensity:  optimalWindow.AvgIntensity,
		Immediate:          immediate,
		CarbonSavings:      carbonSavings,
		AlternativeWindows: alternativeWindows,
	}, nil
}

// findOptimalWindow uses sliding window algorithm to find lowest carbon window
func (s *CarbonScheduler) findOptimalWindow(forecast []carbon.CarbonIntensity, duration time.Duration, minStart, deadline time.Time) (TimeWindow, []TimeWindow) {
	// Convert forecast to time-series data structure
	slots := s.buildTimeSlots(forecast, minStart, deadline)

	// Calculate window size in slots
	windowSlots := int(math.Ceil(float64(duration) / float64(s.slotDuration)))

	if windowSlots > len(slots) {
		// Job duration exceeds forecast range - use entire range
		avgIntensity := s.calculateAverageIntensity(slots)
		return TimeWindow{
			StartTime:    slots[0].Timestamp,
			EndTime:      slots[len(slots)-1].Timestamp.Add(s.slotDuration),
			AvgIntensity: avgIntensity,
			CarbonCost:   avgIntensity * duration.Hours(),
		}, nil
	}

	// Sliding window algorithm
	var optimalWindow TimeWindow
	var alternativeWindows []TimeWindow
	minIntensity := math.MaxFloat64

	for i := 0; i <= len(slots)-windowSlots; i++ {
		windowEnd := i + windowSlots
		windowSlice := slots[i:windowEnd]

		// Calculate average intensity for this window
		avgIntensity := s.calculateAverageIntensity(windowSlice)
		carbonCost := avgIntensity * duration.Hours()

		window := TimeWindow{
			StartTime:    windowSlice[0].Timestamp,
			EndTime:      windowSlice[len(windowSlice)-1].Timestamp.Add(s.slotDuration),
			AvgIntensity: avgIntensity,
			CarbonCost:   carbonCost,
		}

		// Track optimal window
		if avgIntensity < minIntensity {
			minIntensity = avgIntensity
			optimalWindow = window
			alternativeWindows = []TimeWindow{} // Reset alternatives
		} else if math.Abs(avgIntensity-minIntensity) < 10.0 {
			// Track near-optimal windows (within 10 gCO2eq/kWh)
			alternativeWindows = append(alternativeWindows, window)
		}
	}

	// Limit alternative windows to top 3
	if len(alternativeWindows) > 3 {
		alternativeWindows = alternativeWindows[:3]
	}

	return optimalWindow, alternativeWindows
}

// buildTimeSlots converts forecast data into time slots
func (s *CarbonScheduler) buildTimeSlots(forecast []carbon.CarbonIntensity, minStart, deadline time.Time) []carbon.CarbonIntensity {
	var slots []carbon.CarbonIntensity

	for _, point := range forecast {
		// Filter by time constraints
		if point.Timestamp.Before(minStart) || point.Timestamp.After(deadline) {
			continue
		}
		slots = append(slots, point)
	}

	return slots
}

// calculateAverageIntensity computes average carbon intensity for time slots
func (s *CarbonScheduler) calculateAverageIntensity(slots []carbon.CarbonIntensity) float64 {
	if len(slots) == 0 {
		return 0
	}

	sum := 0.0
	for _, slot := range slots {
		sum += slot.Intensity
	}

	return sum / float64(len(slots))
}

// validateRequest checks if scheduling request is valid
func (s *CarbonScheduler) validateRequest(req *ScheduleRequest) error {
	if req.Region == "" {
		return fmt.Errorf("region is required")
	}

	if req.Duration <= 0 {
		return fmt.Errorf("duration must be positive")
	}

	if !req.Deadline.IsZero() && req.Deadline.Before(time.Now()) {
		return fmt.Errorf("deadline must be in the future")
	}

	if !req.MinStartTime.IsZero() && !req.Deadline.IsZero() {
		if req.MinStartTime.Add(req.Duration).After(req.Deadline) {
			return fmt.Errorf("not enough time between min start time and deadline")
		}
	}

	return nil
}

// SetThreshold updates the carbon intensity threshold for immediate execution
func (s *CarbonScheduler) SetThreshold(threshold float64) {
	s.threshold = threshold
}

// SetSlotDuration updates the duration of each time slot
func (s *CarbonScheduler) SetSlotDuration(duration time.Duration) {
	s.slotDuration = duration
}

// ShouldSchedule is a quick check to determine if scheduling is beneficial
func (s *CarbonScheduler) ShouldSchedule(ctx context.Context, region string) (bool, error) {
	current, err := s.fetcher.GetCurrentCarbonIntensity(ctx, region)
	if err != nil {
		return false, err
	}

	// If current intensity is above threshold, scheduling is likely beneficial
	return current.Intensity > s.threshold, nil
}
