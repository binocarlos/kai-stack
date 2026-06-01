package migrations

import "context"

// Jobs creates the background-job queue table. The worker claims rows with
// SELECT ... FOR UPDATE SKIP LOCKED, so this is all the infrastructure the
// queue needs - no extensions, no LISTEN/NOTIFY, works on any pooler mode.
func Jobs(ctx context.Context, m Migrator) error {
	return m.ExecSQL(ctx, `
CREATE TABLE IF NOT EXISTS jobs (
    id           varchar(36) PRIMARY KEY,
    kind         varchar(255) NOT NULL,
    payload      jsonb NOT NULL DEFAULT '{}',
    status       varchar(20) NOT NULL DEFAULT 'pending', -- pending | running | completed | failed
    attempts     int  NOT NULL DEFAULT 0,
    max_attempts int  NOT NULL DEFAULT 3,
    run_at       timestamptz NOT NULL DEFAULT now(),
    last_error   text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

-- Supports the claim query: WHERE status='pending' AND run_at<=now() ORDER BY run_at.
CREATE INDEX IF NOT EXISTS idx_jobs_claim ON jobs (status, run_at);
`)
}
