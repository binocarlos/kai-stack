package jobqueue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/binocarlos/kai-stack/api/pkg/store"
	"github.com/binocarlos/kai-stack/api/pkg/types"
	"github.com/rs/zerolog/log"
)

// Start launches the worker poll loops and returns immediately; the caller
// keeps the process alive (and cancels ctx to stop). Each of the configured
// concurrency slots runs its own loop, claiming jobs via SKIP LOCKED so they
// never collide - including across multiple worker replicas.
func (c *Client) Start(ctx context.Context) error {
	workers := c.cfg.Worker.Concurrency
	if workers < 1 {
		workers = 1
	}
	for i := 0; i < workers; i++ {
		go c.pollLoop(ctx, i)
	}
	log.Info().Int("workers", workers).Msg("job queue workers started")
	return nil
}

// pollLoop claims and runs jobs until the context is cancelled, sleeping for
// PollInterval whenever the queue is empty or errors.
func (c *Client) pollLoop(ctx context.Context, id int) {
	interval := c.cfg.Worker.PollInterval
	if interval <= 0 {
		interval = time.Second
	}

	for {
		if ctx.Err() != nil {
			return
		}

		job, err := c.store.Jobs().Claim(ctx)
		if err != nil {
			if !errors.Is(err, store.ErrNoJob) {
				log.Error().Err(err).Int("worker", id).Msg("failed to claim job")
			}
			if !sleep(ctx, interval) {
				return
			}
			continue
		}

		c.process(ctx, job)
	}
}

// process dispatches a claimed job to its handler and records the outcome.
func (c *Client) process(ctx context.Context, job *types.Job) {
	handler, ok := c.handlers[job.Kind]
	if !ok {
		err := fmt.Errorf("no handler registered for job kind %q", job.Kind)
		log.Error().Err(err).Str("job_id", job.ID).Msg("dropping job")
		_ = c.store.Jobs().Fail(ctx, job, err, 0)
		return
	}

	if err := handler(ctx, job.Payload); err != nil {
		log.Error().Err(err).
			Str("job_id", job.ID).Str("kind", job.Kind).Int("attempt", job.Attempts).
			Msg("job failed")
		_ = c.store.Jobs().Fail(ctx, job, err, backoff(job.Attempts))
		return
	}

	if err := c.store.Jobs().Complete(ctx, job.ID); err != nil {
		log.Error().Err(err).Str("job_id", job.ID).Msg("failed to mark job complete")
	}
}

// backoff is a simple linear retry delay: attempt 1 -> 5s, 2 -> 10s, ...
func backoff(attempts int) time.Duration {
	return time.Duration(attempts) * 5 * time.Second
}

// sleep waits for d or until ctx is cancelled. It reports false if cancelled.
func sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
