package cleanup

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type CleanupService struct {
	db     store.Querier
	logger *logger.Logger
	config Config
}

type Config struct {
	// How often to run cleanup
	Interval time.Duration
	// Delete completed jobs older than this
	CompletedJobRetention time.Duration
	// Delete failed jobs older than this
	FailedJobRetention time.Duration
	// Mark workers unhealthy if not seen in this duration
	WorkerHealthTimeout time.Duration
}

func DefaultConfig() Config {
	return Config{
		Interval:              30 * time.Second,    // Run cleanup frequently
		CompletedJobRetention: 7 * 24 * time.Hour,  // 7 days
		FailedJobRetention:    30 * 24 * time.Hour, // 30 days
		WorkerHealthTimeout:   30 * time.Second,    // Mark unhealthy after 30s
	}
}

func New(db store.Querier, l *logger.Logger, cfg Config) *CleanupService {
	return &CleanupService{
		db:     db,
		logger: l,
		config: cfg,
	}
}

func (c *CleanupService) Start(ctx context.Context) error {
	ticker := time.NewTicker(c.config.Interval)
	defer ticker.Stop()

	// Run once immediately
	c.runCleanup(ctx)

	c.logger.Info("Cleanup service started",
		"interval", c.config.Interval,
		"job_retention", c.config.CompletedJobRetention)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			c.runCleanup(ctx)
		}
	}
}

func (c *CleanupService) runCleanup(ctx context.Context) {
	c.logger.Debug("Running cleanup cycle")

	// 1. Mark stale workers as unhealthy
	workers, err := c.db.MarkWorkersUnhealthy(ctx)
	if err != nil {
		c.logger.Error("Failed to mark workers unhealthy", "error", err)
	} else {
		for _, w := range workers {
			c.logger.Warn("Marked worker unhealthy", "worker_id", w.ID, "hostname", w.Hostname)

			// Create notification (System alert)
			_, err := c.db.CreateNotification(ctx, store.CreateNotificationParams{
				UserID:  pgtype.UUID{Valid: false}, // System notification
				Title:   "Worker Offline",
				Message: fmt.Sprintf("Worker '%s' (%s) has gone offline.", w.Hostname, w.ID),
				Type:    pgtype.Text{String: "error", Valid: true},
				Link:    pgtype.Text{String: "/workers", Valid: true},
			})
			if err != nil {
				c.logger.Error("Failed to create worker notification", "error", err)
			}
		}
	}

	// 2. Clean up old completed jobs
	completedCount, err := c.cleanupOldJobs(ctx, store.JobStatusCompleted, c.config.CompletedJobRetention)
	if err != nil {
		c.logger.Error("Failed to cleanup completed jobs", "error", err)
	} else if completedCount > 0 {
		c.logger.Info("Cleaned up old completed jobs", "count", completedCount)
	}

	// 3. Clean up old failed jobs
	failedCount, err := c.cleanupOldJobs(ctx, store.JobStatusFailed, c.config.FailedJobRetention)
	if err != nil {
		c.logger.Error("Failed to cleanup failed jobs", "error", err)
	} else if failedCount > 0 {
		c.logger.Info("Cleaned up old failed jobs", "count", failedCount)
	}

	// 4. Clean up orphaned tasks (tasks without jobs)
	orphanCount, err := c.cleanupOrphanedTasks(ctx)
	if err != nil {
		c.logger.Error("Failed to cleanup orphaned tasks", "error", err)
	} else if orphanCount > 0 {
		c.logger.Info("Cleaned up orphaned tasks", "count", orphanCount)
	}

	// 5. Clean up old audit logs (optional - keep 90 days)
	auditCount, err := c.cleanupOldAuditLogs(ctx, 90*24*time.Hour)
	if err != nil {
		c.logger.Error("Failed to cleanup audit logs", "error", err)
	} else if auditCount > 0 {
		c.logger.Info("Cleaned up old audit logs", "count", auditCount)
	}

	// 6. Clean up stale unhealthy workers (keep 5 minutes)
	workerCount, err := c.cleanupOldWorkers(ctx, 5*time.Minute)
	if err != nil {
		c.logger.Error("Failed to cleanup old workers", "error", err)
	} else if workerCount > 0 {
		c.logger.Info("Cleaned up old workers", "count", workerCount)
	}
}

func (c *CleanupService) cleanupOldWorkers(ctx context.Context, retention time.Duration) (int64, error) {
	cutoff := time.Now().Add(-retention)
	return c.db.DeleteOldWorkers(ctx, pgtype.Timestamptz{Time: cutoff, Valid: true})
}

func (c *CleanupService) cleanupOldJobs(ctx context.Context, status store.JobStatus, retention time.Duration) (int64, error) {
	cutoff := time.Now().Add(-retention)
	return c.db.DeleteOldJobsByStatus(ctx, store.DeleteOldJobsByStatusParams{
		Status:    status,
		CreatedAt: pgtype.Timestamptz{Time: cutoff, Valid: true},
	})
}

func (c *CleanupService) cleanupOrphanedTasks(ctx context.Context) (int64, error) {
	return c.db.DeleteOrphanedTasks(ctx)
}

func (c *CleanupService) cleanupOldAuditLogs(ctx context.Context, retention time.Duration) (int64, error) {
	cutoff := time.Now().Add(-retention)
	return c.db.DeleteOldAuditLogs(ctx, pgtype.Timestamptz{Time: cutoff, Valid: true})
}
