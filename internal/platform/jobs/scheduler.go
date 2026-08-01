// Package jobs provides a simple background job scheduler.
// Jobs run on a fixed interval using a ticker. Future implementations
// may support cron expressions, distributed locking, or external job queues.
package jobs

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Job represents a background task that runs periodically.
type Job interface {
	// Name returns a human-readable name for the job (used in logging).
	Name() string
	// Run executes the job. Return an error to log a failure; the scheduler
	// will retry on the next interval regardless.
	Run(ctx context.Context) error
	// Interval returns how often the job should run.
	Interval() time.Duration
}

// Scheduler manages background job execution.
type Scheduler struct {
	logger *slog.Logger
	jobs   []Job
	wg     sync.WaitGroup
}

// NewScheduler creates a new job scheduler.
func NewScheduler(logger *slog.Logger) *Scheduler {
	return &Scheduler{
		logger: logger,
	}
}

// Register adds a job to the scheduler.
func (s *Scheduler) Register(job Job) {
	s.jobs = append(s.jobs, job)
}

// Start begins running all registered jobs in the background.
// Each job runs in its own goroutine with its own ticker.
// Call the returned cancel function or cancel the context to stop all jobs.
func (s *Scheduler) Start(ctx context.Context) {
	for _, j := range s.jobs {
		s.wg.Add(1)
		go s.runJob(ctx, j)
	}
	s.logger.Info("job scheduler started", slog.Int("job_count", len(s.jobs)))
}

// Wait blocks until all jobs have stopped.
func (s *Scheduler) Wait() {
	s.wg.Wait()
}

func (s *Scheduler) runJob(ctx context.Context, job Job) {
	defer s.wg.Done()

	ticker := time.NewTicker(job.Interval())
	defer ticker.Stop()

	s.logger.Info("job registered",
		slog.String("job", job.Name()),
		slog.Duration("interval", job.Interval()),
	)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("job stopping", slog.String("job", job.Name()))
			return
		case <-ticker.C:
			if err := job.Run(ctx); err != nil {
				s.logger.Error("job failed",
					slog.String("job", job.Name()),
					slog.String("error", err.Error()),
				)
			}
		}
	}
}
