package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisQueue handles Redis-based queue operations
type RedisQueue struct {
	client            *redis.Client
	immediateQueueKey string
	delayedSetKey     string
}

// QueueItem represents an item in the queue
type QueueItem struct {
	JobID         string    `json:"job_id"`
	DockerImage   string    `json:"docker_image"`
	Command       *string   `json:"command,omitempty"`
	ScheduledTime time.Time `json:"scheduled_time"`
	Priority      int       `json:"priority"`
}

// NewRedisQueue creates a new Redis queue client
func NewRedisQueue(addr, password string, db int, immediateKey, delayedKey string) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("✓ Successfully connected to Redis")

	return &RedisQueue{
		client:            client,
		immediateQueueKey: immediateKey,
		delayedSetKey:     delayedKey,
	}, nil
}

// Close closes the Redis connection
func (q *RedisQueue) Close() error {
	log.Println("Closing Redis connection...")
	return q.client.Close()
}

// EnqueueImmediate adds a job to the immediate execution queue (FIFO List)
func (q *RedisQueue) EnqueueImmediate(ctx context.Context, item *QueueItem) error {
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal queue item: %w", err)
	}

	// Push to the right end of the list (FIFO)
	if err := q.client.RPush(ctx, q.immediateQueueKey, data).Err(); err != nil {
		return fmt.Errorf("failed to enqueue immediate job: %w", err)
	}

	log.Printf("✓ Enqueued immediate job: %s", item.JobID)
	return nil
}

// EnqueueDelayed adds a job to the delayed execution queue (Sorted Set with timestamp score)
func (q *RedisQueue) EnqueueDelayed(ctx context.Context, item *QueueItem) error {
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal queue item: %w", err)
	}

	// Use scheduled time's Unix timestamp as the score for the sorted set
	score := float64(item.ScheduledTime.Unix())

	member := redis.Z{
		Score:  score,
		Member: data,
	}

	if err := q.client.ZAdd(ctx, q.delayedSetKey, member).Err(); err != nil {
		return fmt.Errorf("failed to enqueue delayed job: %w", err)
	}

	log.Printf("✓ Enqueued delayed job: %s (scheduled for %s)", item.JobID, item.ScheduledTime.Format(time.RFC3339))
	return nil
}

// DequeueImmediate retrieves and removes a job from the immediate queue
func (q *RedisQueue) DequeueImmediate(ctx context.Context) (*QueueItem, error) {
	// Pop from the left end of the list (FIFO)
	result, err := q.client.LPop(ctx, q.immediateQueueKey).Result()
	if err == redis.Nil {
		return nil, nil // Queue is empty
	}
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue immediate job: %w", err)
	}

	var item QueueItem
	if err := json.Unmarshal([]byte(result), &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal queue item: %w", err)
	}

	return &item, nil
}

// GetDueDelayedJobs retrieves jobs from delayed queue that are due for execution
func (q *RedisQueue) GetDueDelayedJobs(ctx context.Context, limit int64) ([]*QueueItem, error) {
	now := float64(time.Now().Unix())

	// Get jobs with score (timestamp) <= now
	results, err := q.client.ZRangeByScore(ctx, q.delayedSetKey, &redis.ZRangeBy{
		Min:   "-inf",
		Max:   fmt.Sprintf("%f", now),
		Count: limit,
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to get due delayed jobs: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	var items []*QueueItem
	for _, result := range results {
		var item QueueItem
		if err := json.Unmarshal([]byte(result), &item); err != nil {
			log.Printf("Warning: failed to unmarshal delayed job: %v", err)
			continue
		}
		items = append(items, &item)
	}

	return items, nil
}

// RemoveDelayedJob removes a job from the delayed queue
func (q *RedisQueue) RemoveDelayedJob(ctx context.Context, jobID string) error {
	// We need to find and remove by member value
	// First, get all members and find the one with matching jobID
	results, err := q.client.ZRange(ctx, q.delayedSetKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get delayed jobs: %w", err)
	}

	for _, result := range results {
		var item QueueItem
		if err := json.Unmarshal([]byte(result), &item); err != nil {
			continue
		}

		if item.JobID == jobID {
			if err := q.client.ZRem(ctx, q.delayedSetKey, result).Err(); err != nil {
				return fmt.Errorf("failed to remove delayed job: %w", err)
			}
			log.Printf("✓ Removed delayed job: %s", jobID)
			return nil
		}
	}

	return fmt.Errorf("job not found in delayed queue")
}

// GetImmediateQueueLength returns the length of the immediate queue
func (q *RedisQueue) GetImmediateQueueLength(ctx context.Context) (int64, error) {
	length, err := q.client.LLen(ctx, q.immediateQueueKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get immediate queue length: %w", err)
	}
	return length, nil
}

// GetDelayedQueueLength returns the length of the delayed queue
func (q *RedisQueue) GetDelayedQueueLength(ctx context.Context) (int64, error) {
	length, err := q.client.ZCard(ctx, q.delayedSetKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get delayed queue length: %w", err)
	}
	return length, nil
}

// HealthCheck performs a Redis health check
func (q *RedisQueue) HealthCheck(ctx context.Context) error {
	return q.client.Ping(ctx).Err()
}

// GetReadyDelayedJobs retrieves all jobs from delayed queue that are ready to execute
func (q *RedisQueue) GetReadyDelayedJobs(ctx context.Context, now time.Time) ([]*QueueItem, error) {
	score := float64(now.Unix())

	// Get jobs with score (timestamp) <= now
	results, err := q.client.ZRangeByScore(ctx, q.delayedSetKey, &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%f", score),
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to get ready delayed jobs: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	var items []*QueueItem
	for _, result := range results {
		var item QueueItem
		if err := json.Unmarshal([]byte(result), &item); err != nil {
			log.Printf("Warning: failed to unmarshal delayed job: %v", err)
			continue
		}
		items = append(items, &item)
	}

	return items, nil
}

// RemoveFromDelayed removes a specific job from the delayed queue by job ID
func (q *RedisQueue) RemoveFromDelayed(ctx context.Context, jobID string) error {
	// Get all members and find the one with matching jobID
	results, err := q.client.ZRange(ctx, q.delayedSetKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get delayed jobs: %w", err)
	}

	for _, result := range results {
		var item QueueItem
		if err := json.Unmarshal([]byte(result), &item); err != nil {
			continue
		}

		if item.JobID == jobID {
			if err := q.client.ZRem(ctx, q.delayedSetKey, result).Err(); err != nil {
				return fmt.Errorf("failed to remove delayed job: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("job %s not found in delayed queue", jobID)
}

// GetDelayedQueueStats returns statistics about the delayed queue
func (q *RedisQueue) GetDelayedQueueStats(ctx context.Context) (map[string]interface{}, error) {
	totalDelayed, err := q.GetDelayedQueueLength(ctx)
	if err != nil {
		return nil, err
	}

	// Count ready jobs
	readyJobs, err := q.GetReadyDelayedJobs(ctx, time.Now())
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_delayed_jobs": totalDelayed,
		"ready_jobs":         len(readyJobs),
		"pending_jobs":       totalDelayed - int64(len(readyJobs)),
	}

	return stats, nil
}

// GetQueueLength returns the length of the immediate queue (alias for metrics)
func (q *RedisQueue) GetQueueLength(ctx context.Context) (int64, error) {
	return q.GetImmediateQueueLength(ctx)
}

// GetDelayedJobsCount returns the count of delayed jobs (alias for metrics)
func (q *RedisQueue) GetDelayedJobsCount(ctx context.Context) (int64, error) {
	return q.GetDelayedQueueLength(ctx)
}

// SetWorkerHeartbeat sets a worker heartbeat key with expiration
func (q *RedisQueue) SetWorkerHeartbeat(ctx context.Context, workerID string, ttlSeconds int) error {
	key := fmt.Sprintf("worker:%s", workerID)
	return q.client.Set(ctx, key, "alive", time.Duration(ttlSeconds)*time.Second).Err()
}

// GetActiveWorkers scans for active worker keys and returns their IDs
func (q *RedisQueue) GetActiveWorkers(ctx context.Context) ([]string, error) {
	var workers []string

	iter := q.client.Scan(ctx, 0, "worker:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		// Extract worker ID from "worker:{uuid}" format
		if len(key) > 7 {
			workerID := key[7:] // Remove "worker:" prefix
			workers = append(workers, workerID)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan worker keys: %w", err)
	}

	return workers, nil
}
