package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Sambit-Mondal/karbos/server/internal/queue"
	"github.com/Sambit-Mondal/karbos/server/internal/worker"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsCollector handles Prometheus metrics collection
type MetricsCollector struct {
	// Prometheus metrics
	jobsPending    prometheus.Gauge
	jobsRunning    prometheus.Gauge
	co2SavedTotal  prometheus.Counter
	metricsHandler http.Handler

	// Data sources
	queue      *queue.RedisQueue
	workerPool *worker.Pool
	db         *sql.DB

	// Control
	mu      sync.RWMutex
	enabled bool
}

// NewMetricsCollector creates a new Prometheus metrics collector
func NewMetricsCollector(queue *queue.RedisQueue, workerPool *worker.Pool, db *sql.DB) *MetricsCollector {
	// Create Prometheus metrics
	jobsPending := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "karbos_jobs_pending",
		Help: "Number of jobs waiting in queue (immediate + delayed)",
	})

	jobsRunning := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "karbos_jobs_running",
		Help: "Number of jobs currently being executed by workers",
	})

	co2SavedTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "karbos_co2_saved_total_grams",
		Help: "Total grams of CO2 saved through carbon-aware scheduling",
	})

	// Register metrics with Prometheus
	prometheus.MustRegister(jobsPending)
	prometheus.MustRegister(jobsRunning)
	prometheus.MustRegister(co2SavedTotal)

	collector := &MetricsCollector{
		jobsPending:    jobsPending,
		jobsRunning:    jobsRunning,
		co2SavedTotal:  co2SavedTotal,
		metricsHandler: promhttp.Handler(),
		queue:          queue,
		workerPool:     workerPool,
		db:             db,
		enabled:        true,
	}

	log.Println("✓ Prometheus metrics collector initialized")
	return collector
}

// UpdateMetrics refreshes all metrics from their data sources
func (m *MetricsCollector) UpdateMetrics(ctx context.Context) error {
	m.mu.RLock()
	if !m.enabled {
		m.mu.RUnlock()
		return nil
	}
	m.mu.RUnlock()

	// Update jobs_pending (queue depth)
	if err := m.updateJobsPending(ctx); err != nil {
		log.Printf("Warning: Failed to update jobs_pending metric: %v", err)
	}

	// Update jobs_running (active containers)
	// Skip if worker pool not configured (e.g., API server)
	if err := m.updateJobsRunning(); err != nil {
		// Only log if it's not a "worker pool not configured" error
		if m.workerPool != nil {
			log.Printf("Warning: Failed to update jobs_running metric: %v", err)
		}
	}

	// Update co2_saved_total (cumulative savings)
	if err := m.updateCO2Saved(ctx); err != nil {
		log.Printf("Warning: Failed to update co2_saved_total metric: %v", err)
	}

	return nil
}

// updateJobsPending counts jobs in both immediate and delayed queues
func (m *MetricsCollector) updateJobsPending(ctx context.Context) error {
	if m.queue == nil {
		return fmt.Errorf("queue not configured")
	}

	// Get immediate queue length
	immediateLen, err := m.queue.GetQueueLength(ctx)
	if err != nil {
		return fmt.Errorf("failed to get immediate queue length: %w", err)
	}

	// Get delayed queue length
	delayedLen, err := m.queue.GetDelayedJobsCount(ctx)
	if err != nil {
		return fmt.Errorf("failed to get delayed queue length: %w", err)
	}

	total := float64(immediateLen + delayedLen)
	m.jobsPending.Set(total)

	return nil
}

// updateJobsRunning counts currently active jobs from worker pool
func (m *MetricsCollector) updateJobsRunning() error {
	if m.workerPool == nil {
		return fmt.Errorf("worker pool not configured")
	}

	activeCount := m.workerPool.GetActiveJobCount()
	m.jobsRunning.Set(float64(activeCount))

	return nil
}

// updateCO2Saved calculates total CO2 savings from completed jobs
func (m *MetricsCollector) updateCO2Saved(ctx context.Context) error {
	if m.db == nil {
		return fmt.Errorf("database not configured")
	}

	// Calculate CO2 savings based on execution logs
	// Estimation: average power usage (50W) * duration * carbon intensity difference
	// For now, we'll track completed jobs count as a proxy
	// TODO: Implement actual CO2 calculation based on carbon intensity data
	query := `
		SELECT COUNT(*) as completed_jobs
		FROM jobs
		WHERE status = 'COMPLETED'
	`

	var completedJobs int
	err := m.db.QueryRowContext(ctx, query).Scan(&completedJobs)
	if err != nil {
		if err == sql.ErrNoRows {
			completedJobs = 0
		} else {
			return fmt.Errorf("failed to query completed jobs: %w", err)
		}
	}

	// Estimate CO2 saved: assume each job saves ~100g CO2 on average
	// This is a placeholder - actual calculation would need:
	// - Job execution duration from execution_logs
	// - Carbon intensity at scheduling time vs execution time
	// - Estimated power consumption
	estimatedCO2Saved := float64(completedJobs) * 100.0

	// Note: This is cumulative, so we set it directly
	// In a real implementation, we'd track incremental changes
	m.co2SavedTotal.Add(estimatedCO2Saved)

	return nil
}

// ServeHTTP handles the /metrics endpoint
func (m *MetricsCollector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	enabled := m.enabled
	m.mu.RUnlock()

	if !enabled {
		http.Error(w, "Metrics collection is disabled", http.StatusServiceUnavailable)
		return
	}

	// Update metrics before serving
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := m.UpdateMetrics(ctx); err != nil {
		log.Printf("Warning: Failed to update metrics: %v", err)
		// Continue serving stale metrics rather than failing
	}

	// Serve Prometheus metrics
	m.metricsHandler.ServeHTTP(w, r)
}

// StartBackgroundUpdater starts a goroutine that periodically updates metrics
func (m *MetricsCollector) StartBackgroundUpdater(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("✓ Metrics background updater started (interval: %v)", interval)

		for {
			select {
			case <-ctx.Done():
				log.Println("Metrics background updater stopped")
				return
			case <-ticker.C:
				if err := m.UpdateMetrics(ctx); err != nil {
					log.Printf("Warning: Metrics update failed: %v", err)
				}
			}
		}
	}()
}

// Enable enables metrics collection
func (m *MetricsCollector) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
	log.Println("✓ Metrics collection enabled")
}

// Disable disables metrics collection
func (m *MetricsCollector) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
	log.Println("✓ Metrics collection disabled")
}

// IsEnabled returns true if metrics collection is enabled
func (m *MetricsCollector) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// GetMetricsSnapshot returns current metric values (for testing/debugging)
func (m *MetricsCollector) GetMetricsSnapshot(ctx context.Context) (map[string]float64, error) {
	if err := m.UpdateMetrics(ctx); err != nil {
		return nil, err
	}

	// Note: In production, you'd use prometheus client to get current values
	// For now, we'll query the sources directly
	immediateLen, _ := m.queue.GetQueueLength(ctx)
	delayedLen, _ := m.queue.GetDelayedJobsCount(ctx)
	activeJobs := m.workerPool.GetActiveJobCount()

	return map[string]float64{
		"jobs_pending": float64(immediateLen + delayedLen),
		"jobs_running": float64(activeJobs),
		// co2_saved_total would need to be queried from database
	}, nil
}

// GetPrometheusText returns metrics in Prometheus text format for Fiber
func (m *MetricsCollector) GetPrometheusText() string {
	// Use prometheus gatherer to collect metrics
	gatherer := prometheus.DefaultGatherer
	metrics, err := gatherer.Gather()
	if err != nil {
		return fmt.Sprintf("# Error gathering metrics: %v\n", err)
	}

	// Format as Prometheus text
	var result string
	for _, metric := range metrics {
		// Only include our karbos metrics
		if metric.GetName() == "karbos_jobs_pending" ||
			metric.GetName() == "karbos_jobs_running" ||
			metric.GetName() == "karbos_co2_saved_total_grams" {
			result += fmt.Sprintf("# HELP %s %s\n", metric.GetName(), metric.GetHelp())
			result += fmt.Sprintf("# TYPE %s %s\n", metric.GetName(), metric.GetType())
			for _, m := range metric.GetMetric() {
				if m.GetGauge() != nil {
					result += fmt.Sprintf("%s %f\n", metric.GetName(), m.GetGauge().GetValue())
				} else if m.GetCounter() != nil {
					result += fmt.Sprintf("%s %f\n", metric.GetName(), m.GetCounter().GetValue())
				}
			}
		}
	}
	return result
}
