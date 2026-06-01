package migrations

import (
	"context"
	"database/sql"
)

// Migrator is the minimal database surface a migration is given. It is
// satisfied by a transaction-scoped adapter in the store package, so every
// migration runs atomically and is tracked in the same transaction.
type Migrator interface {
	ExecSQL(ctx context.Context, sql string, args ...any) error
	QuerySQL(ctx context.Context, sql string, args ...any) (*sql.Rows, error)
}

// MigrationFunc applies a single migration's forward (Up) change.
type MigrationFunc func(ctx context.Context, m Migrator) error

// Migration pairs a stable, unique name with its forward function. The name
// is what gets recorded in the go_schema_migrations ledger.
type Migration struct {
	Name string
	Up   MigrationFunc
}
