package migrations

import "context"

// InitialSchema creates the example_records table. It matches the
// types.ExampleRecord GORM model (timestamps are int64 unix millis, config is
// JSONB) so the Go struct and the SQL agree.
func InitialSchema(ctx context.Context, m Migrator) error {
	return m.ExecSQL(ctx, `
CREATE TABLE IF NOT EXISTS example_records (
    id         varchar(36) PRIMARY KEY,
    user_id    varchar(36) NOT NULL,
    created_at bigint NOT NULL DEFAULT 0,
    updated_at bigint NOT NULL DEFAULT 0,
    config     jsonb
);

CREATE INDEX IF NOT EXISTS idx_example_records_user_id
    ON example_records (user_id);
`)
}
