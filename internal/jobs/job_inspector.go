package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// JobInspector provides methods to inspect and manage background jobs
type JobInspector interface {
	ListAllJobs(ctx context.Context) ([]JobInfo, map[string]int, error)
	GetJobInfo(ctx context.Context, jobID string) (*JobInfo, error)
	RetryJob(ctx context.Context, jobID string) error
	CancelJob(ctx context.Context, jobID string) error
	GetQueueStats(ctx context.Context) (map[string]int, error)
}

// JobInfo represents detailed information about a job
type JobInfo struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Payload     map[string]interface{} `json:"payload"`
	State       string                 `json:"state"` // pending, active, completed, failed, archived
	Queue       string                 `json:"queue"`
	MaxRetry    int                    `json:"max_retry"`
	Retried     int                    `json:"retried"`
	LastErr     string                 `json:"last_err,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// AsynqJobInspector implements JobInspector using Asynq Inspector
type AsynqJobInspector struct {
	inspector *asynq.Inspector
}

// NewAsynqJobInspector creates a new Asynq-based job inspector
func NewAsynqJobInspector(redisAddr, redisPassword string, redisDB int) *AsynqJobInspector {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	return &AsynqJobInspector{
		inspector: inspector,
	}
}

// ListAllJobs retrieves jobs from all states
func (i *AsynqJobInspector) ListAllJobs(ctx context.Context) ([]JobInfo, map[string]int, error) {
	queues, err := i.inspector.Queues()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get queues: %w", err)
	}

	var allJobs []JobInfo
	stats := map[string]int{
		"total":     0,
		"pending":   0,
		"active":    0,
		"completed": 0,
		"failed":    0,
		"archived":  0,
	}

	// Get jobs from each queue
	for _, queueName := range queues {
		// Pending tasks
		pending, err := i.inspector.ListPendingTasks(queueName, asynq.PageSize(100))
		if err == nil {
			for _, task := range pending {
				job := convertTaskToJobInfo(task, "pending", queueName)
				allJobs = append(allJobs, job)
				stats["pending"]++
				stats["total"]++
			}
		}

		// Active tasks
		active, err := i.inspector.ListActiveTasks(queueName, asynq.PageSize(100))
		if err == nil {
			for _, task := range active {
				job := convertTaskToJobInfo(task, "active", queueName)
				allJobs = append(allJobs, job)
				stats["active"]++
				stats["total"]++
			}
		}

		// Scheduled tasks (treat as pending)
		scheduled, err := i.inspector.ListScheduledTasks(queueName, asynq.PageSize(100))
		if err == nil {
			for _, task := range scheduled {
				job := convertTaskToJobInfo(task, "pending", queueName)
				allJobs = append(allJobs, job)
				stats["pending"]++
				stats["total"]++
			}
		}

		// Retry tasks (failed but will retry)
		retry, err := i.inspector.ListRetryTasks(queueName, asynq.PageSize(100))
		if err == nil {
			for _, task := range retry {
				job := convertTaskToJobInfo(task, "failed", queueName)
				allJobs = append(allJobs, job)
				stats["failed"]++
				stats["total"]++
			}
		}

		// Archived tasks (permanently failed)
		archived, err := i.inspector.ListArchivedTasks(queueName, asynq.PageSize(100))
		if err == nil {
			for _, task := range archived {
				job := convertTaskToJobInfo(task, "archived", queueName)
				allJobs = append(allJobs, job)
				stats["archived"]++
				stats["total"]++
			}
		}

		// Completed tasks
		completed, err := i.inspector.ListCompletedTasks(queueName, asynq.PageSize(100))
		if err == nil {
			for _, task := range completed {
				job := convertTaskToJobInfo(task, "completed", queueName)
				allJobs = append(allJobs, job)
				stats["completed"]++
				stats["total"]++
			}
		}
	}

	return allJobs, stats, nil
}

// GetJobInfo retrieves detailed information about a specific job
func (i *AsynqJobInspector) GetJobInfo(ctx context.Context, jobID string) (*JobInfo, error) {
	queues, err := i.inspector.Queues()
	if err != nil {
		return nil, fmt.Errorf("failed to get queues: %w", err)
	}

	// Search for the job in all queues and states
	for _, queueName := range queues {
		// Check pending
		pending, err := i.inspector.ListPendingTasks(queueName)
		if err == nil {
			for _, task := range pending {
				if task.ID == jobID {
					job := convertTaskToJobInfo(task, "pending", queueName)
					return &job, nil
				}
			}
		}

		// Check active
		active, err := i.inspector.ListActiveTasks(queueName)
		if err == nil {
			for _, task := range active {
				if task.ID == jobID {
					job := convertTaskToJobInfo(task, "active", queueName)
					return &job, nil
				}
			}
		}

		// Check scheduled
		scheduled, err := i.inspector.ListScheduledTasks(queueName)
		if err == nil {
			for _, task := range scheduled {
				if task.ID == jobID {
					job := convertTaskToJobInfo(task, "pending", queueName)
					return &job, nil
				}
			}
		}

		// Check retry
		retry, err := i.inspector.ListRetryTasks(queueName)
		if err == nil {
			for _, task := range retry {
				if task.ID == jobID {
					job := convertTaskToJobInfo(task, "failed", queueName)
					return &job, nil
				}
			}
		}

		// Check archived
		archived, err := i.inspector.ListArchivedTasks(queueName)
		if err == nil {
			for _, task := range archived {
				if task.ID == jobID {
					job := convertTaskToJobInfo(task, "archived", queueName)
					return &job, nil
				}
			}
		}

		// Check completed
		completed, err := i.inspector.ListCompletedTasks(queueName)
		if err == nil {
			for _, task := range completed {
				if task.ID == jobID {
					job := convertTaskToJobInfo(task, "completed", queueName)
					return &job, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("job not found: %s", jobID)
}

// RetryJob queues a failed job for retry using Asynq's RunTask
func (i *AsynqJobInspector) RetryJob(ctx context.Context, jobID string) error {
	queues, err := i.inspector.Queues()
	if err != nil {
		return fmt.Errorf("failed to get queues: %w", err)
	}

	// Try to run the task in each queue
	for _, queueName := range queues {
		// RunTask moves task from scheduled/retry/archived to pending
		if err := i.inspector.RunTask(queueName, jobID); err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to retry job: %s (job not found or not in retryable state)", jobID)
}

// CancelJob cancels a pending or active job
func (i *AsynqJobInspector) CancelJob(ctx context.Context, jobID string) error {
	queues, err := i.inspector.Queues()
	if err != nil {
		return fmt.Errorf("failed to get queues: %w", err)
	}

	for _, queueName := range queues {
		// Try to delete from pending
		if err := i.inspector.DeleteTask(queueName, jobID); err == nil {
			return nil
		}

		// Try to archive active task (Asynq doesn't support killing active tasks directly)
		if err := i.inspector.ArchiveTask(queueName, jobID); err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to cancel job: %s", jobID)
}

// GetQueueStats returns statistics for all queues
func (i *AsynqJobInspector) GetQueueStats(ctx context.Context) (map[string]int, error) {
	queues, err := i.inspector.Queues()
	if err != nil {
		return nil, fmt.Errorf("failed to get queues: %w", err)
	}

	totalStats := map[string]int{
		"total":     0,
		"pending":   0,
		"active":    0,
		"completed": 0,
		"failed":    0,
		"archived":  0,
	}

	for _, queueName := range queues {
		queueStats, err := i.inspector.GetQueueInfo(queueName)
		if err != nil {
			continue
		}

		totalStats["pending"] += queueStats.Pending + queueStats.Scheduled
		totalStats["active"] += queueStats.Active
		totalStats["completed"] += queueStats.Completed
		totalStats["failed"] += queueStats.Retry
		totalStats["archived"] += queueStats.Archived
	}

	totalStats["total"] = totalStats["pending"] + totalStats["active"] + 
		totalStats["completed"] + totalStats["failed"] + totalStats["archived"]

	return totalStats, nil
}

// Close closes the inspector connection
func (i *AsynqJobInspector) Close() error {
	return i.inspector.Close()
}

// convertTaskToJobInfo converts an Asynq TaskInfo to our JobInfo struct
func convertTaskToJobInfo(task *asynq.TaskInfo, state, queue string) JobInfo {
	var payload map[string]interface{}
	// Asynq payload is raw bytes, we store it as-is for now
	// In a real implementation, you'd unmarshal based on task type
	payload = map[string]interface{}{
		"raw": string(task.Payload),
	}

	lastErr := ""
	if task.LastErr != "" {
		lastErr = task.LastErr
	}

	var processedAt *time.Time
	var completedAt *time.Time
	if state == "completed" || state == "failed" || state == "archived" {
		now := time.Now()
		completedAt = &now
	}

	return JobInfo{
		ID:          task.ID,
		Type:        task.Type,
		Payload:     payload,
		State:       state,
		Queue:       queue,
		MaxRetry:    task.MaxRetry,
		Retried:     task.Retried,
		LastErr:     lastErr,
		CreatedAt:   time.Now(), // Asynq doesn't expose created_at, use current time
		ProcessedAt: processedAt,
		CompletedAt: completedAt,
	}
}
