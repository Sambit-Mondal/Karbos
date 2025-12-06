package docker

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Service handles Docker container operations
type Service struct {
	client *client.Client
}

// ContainerResult holds the output and metadata from container execution
type ContainerResult struct {
	Output    string
	ExitCode  int
	Duration  int // in seconds
	StartedAt time.Time
	Error     error
}

// NewDockerService creates a new Docker service instance
func NewDockerService() (*Service, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &Service{client: cli}, nil
}

// Close closes the Docker client connection
func (s *Service) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// Ping checks if Docker daemon is accessible
func (s *Service) Ping(ctx context.Context) error {
	_, err := s.client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping Docker daemon: %w", err)
	}
	return nil
}

// PullImage pulls a Docker image if not already present
func (s *Service) PullImage(ctx context.Context, imageName string) error {
	// Check if image exists locally
	_, _, err := s.client.ImageInspectWithRaw(ctx, imageName)
	if err == nil {
		// Image already exists
		return nil
	}

	// Pull the image
	reader, err := s.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// Wait for pull to complete (discard output for now)
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("failed to read pull response: %w", err)
	}

	return nil
}

// RunContainer runs a Docker container and captures its output
// This is the main function that executes user code
func (s *Service) RunContainer(ctx context.Context, imageName string, command []string) (*ContainerResult, error) {
	result := &ContainerResult{
		StartedAt: time.Now(),
	}

	// Pull image if needed
	if err := s.PullImage(ctx, imageName); err != nil {
		result.Error = err
		return result, err
	}

	// Create container configuration
	containerConfig := &container.Config{
		Image:        imageName,
		Cmd:          command,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}

	// Host configuration (resource limits, etc.)
	hostConfig := &container.HostConfig{
		AutoRemove: false, // We'll remove manually after capturing logs
		Resources: container.Resources{
			// Add resource limits to prevent abuse
			Memory:     512 * 1024 * 1024, // 512MB
			MemorySwap: 512 * 1024 * 1024, // No swap
			CPUQuota:   50000,             // 50% of one CPU
		},
	}

	// Create container
	resp, err := s.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		result.Error = fmt.Errorf("failed to create container: %w", err)
		return result, result.Error
	}
	containerID := resp.ID

	// Ensure cleanup
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.client.ContainerRemove(cleanupCtx, containerID, container.RemoveOptions{
			Force: true,
		})
	}()

	// Start container
	if err := s.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		result.Error = fmt.Errorf("failed to start container: %w", err)
		return result, result.Error
	}

	// Wait for container to finish
	statusCh, errCh := s.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			result.Error = fmt.Errorf("error waiting for container: %w", err)
			return result, result.Error
		}
	case status := <-statusCh:
		result.ExitCode = int(status.StatusCode)
	case <-ctx.Done():
		result.Error = fmt.Errorf("context cancelled while waiting for container")
		return result, result.Error
	}

	// Capture logs
	logOptions := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: false,
		Follow:     false,
	}

	logs, err := s.client.ContainerLogs(ctx, containerID, logOptions)
	if err != nil {
		result.Error = fmt.Errorf("failed to get container logs: %w", err)
		return result, result.Error
	}
	defer logs.Close()

	// Read stdout and stderr
	var stdout, stderr strings.Builder
	_, err = stdcopy.StdCopy(&stdout, &stderr, logs)
	if err != nil {
		result.Error = fmt.Errorf("failed to read container logs: %w", err)
		return result, result.Error
	}

	// Combine stdout and stderr
	if stdout.Len() > 0 {
		result.Output = stdout.String()
	}
	if stderr.Len() > 0 {
		if result.Output != "" {
			result.Output += "\n--- STDERR ---\n"
		}
		result.Output += stderr.String()
	}

	// Calculate duration
	result.Duration = int(time.Since(result.StartedAt).Seconds())

	return result, nil
}

// ListRunningContainers returns the count of currently running containers
func (s *Service) ListRunningContainers(ctx context.Context) (int, error) {
	containers, err := s.client.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to list containers: %w", err)
	}
	return len(containers), nil
}

// GetDockerInfo returns Docker system information
func (s *Service) GetDockerInfo(ctx context.Context) (map[string]interface{}, error) {
	info, err := s.client.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker info: %w", err)
	}

	return map[string]interface{}{
		"containers":         info.Containers,
		"containers_running": info.ContainersRunning,
		"containers_paused":  info.ContainersPaused,
		"containers_stopped": info.ContainersStopped,
		"images":             info.Images,
		"driver":             info.Driver,
		"memory_total":       info.MemTotal,
		"cpus":               info.NCPU,
		"server_version":     info.ServerVersion,
	}, nil
}
