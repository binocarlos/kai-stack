package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/binocarlos/kai-stack/api/pkg/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNoJob is returned by Claim when no runnable job is available.
var ErrNoJob = errors.New("no job available")

// JobRepository is the SKIP LOCKED background-job queue. It is just another
// repository over a plain Postgres table - the worker (pkg/jobqueue) polls it.
type JobRepository struct {
	*Repository[types.Job]
}

func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{
		Repository: NewRepository[types.Job](db),
	}
}

// Enqueue inserts a new pending job. payload is JSON-encoded; pass any
// serialisable value (or nil for an empty payload).
func (r *JobRepository) Enqueue(ctx context.Context, kind string, payload any) (*types.Job, error) {
	data := []byte("{}")
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal job payload: %w", err)
		}
		data = encoded
	}

	job := types.Job{ID: uuid.New().String(), Kind: kind, Payload: data, Status: "pending"}
	if err := r.db.WithContext(ctx).Exec(`
INSERT INTO jobs (id, kind, payload, status)
VALUES (?, ?, ?, 'pending')`,
		job.ID, job.Kind, string(job.Payload),
	).Error; err != nil {
		return nil, fmt.Errorf("failed to enqueue job: %w", err)
	}
	return &job, nil
}

// Claim atomically picks the oldest runnable pending job, marks it running, and
// returns it. Concurrent workers never claim the same row thanks to
// FOR UPDATE SKIP LOCKED. Returns ErrNoJob when the queue is empty.
func (r *JobRepository) Claim(ctx context.Context) (*types.Job, error) {
	var job types.Job
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := tx.Raw(`
SELECT id, kind, payload, status, attempts, max_attempts, COALESCE(last_error, '')
FROM jobs
WHERE status = 'pending' AND run_at <= now()
ORDER BY run_at
FOR UPDATE SKIP LOCKED
LIMIT 1`).Row()

		if err := row.Scan(
			&job.ID, &job.Kind, &job.Payload, &job.Status,
			&job.Attempts, &job.MaxAttempts, &job.LastError,
		); err != nil {
			return err // sql.ErrNoRows when nothing to claim
		}

		return tx.Exec(`
UPDATE jobs SET status = 'running', attempts = attempts + 1, updated_at = now()
WHERE id = ?`, job.ID).Error
	})

	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNoJob
	}
	if err != nil {
		return nil, err
	}
	job.Attempts++ // reflect the increment we just committed
	return &job, nil
}

// Complete marks a job finished successfully.
func (r *JobRepository) Complete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Exec(`
UPDATE jobs SET status = 'completed', updated_at = now() WHERE id = ?`, id).Error
}

// Fail records a failed attempt. If the job still has attempts left it is
// re-queued to run after backoff; otherwise it is marked failed.
func (r *JobRepository) Fail(ctx context.Context, job *types.Job, cause error, backoff time.Duration) error {
	if job.Attempts < job.MaxAttempts {
		runAt := time.Now().Add(backoff)
		return r.db.WithContext(ctx).Exec(`
UPDATE jobs SET status = 'pending', run_at = ?, last_error = ?, updated_at = now()
WHERE id = ?`, runAt, cause.Error(), job.ID).Error
	}
	return r.db.WithContext(ctx).Exec(`
UPDATE jobs SET status = 'failed', last_error = ?, updated_at = now()
WHERE id = ?`, cause.Error(), job.ID).Error
}
