package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/binocarlos/kai-stack/api/pkg/store/migrations"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// txMigrator adapts a GORM transaction to the migrations.Migrator interface so
// each migration runs against the same transaction that records it.
type txMigrator struct {
	tx *gorm.DB
}

func (m *txMigrator) ExecSQL(ctx context.Context, sql string, args ...any) error {
	return m.tx.WithContext(ctx).Exec(sql, args...).Error
}

func (m *txMigrator) QuerySQL(ctx context.Context, sql string, args ...any) (*sql.Rows, error) {
	return m.tx.WithContext(ctx).Raw(sql, args...).Rows()
}

// runMigrations applies any not-yet-applied migrations from migrations.All, in
// order. State is tracked in go_schema_migrations (a table separate from
// Supabase's own supabase_migrations.schema_migrations, so the two never
// collide). Each migration and its ledger insert run in a single transaction:
// on failure the transaction rolls back and the ledger is untouched, making the
// runner safe to re-run.
func (s *PostgresStore) runMigrations(ctx context.Context) error {
	if err := s.gdb.WithContext(ctx).Exec(`
CREATE TABLE IF NOT EXISTS go_schema_migrations (
    name       varchar(255) PRIMARY KEY,
    applied_at timestamptz DEFAULT now()
)`).Error; err != nil {
		return fmt.Errorf("failed to create migrations ledger: %w", err)
	}

	applied, err := s.appliedMigrations(ctx)
	if err != nil {
		return err
	}

	for _, migration := range migrations.All {
		if applied[migration.Name] {
			continue
		}
		if err := s.runMigrationInTx(ctx, migration); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Name, err)
		}
		log.Info().Str("name", migration.Name).Msg("applied migration")
	}

	return nil
}

// appliedMigrations returns the set of migration names already in the ledger.
func (s *PostgresStore) appliedMigrations(ctx context.Context) (map[string]bool, error) {
	var names []string
	if err := s.gdb.WithContext(ctx).
		Raw(`SELECT name FROM go_schema_migrations`).
		Scan(&names).Error; err != nil {
		return nil, fmt.Errorf("failed to read migrations ledger: %w", err)
	}
	applied := make(map[string]bool, len(names))
	for _, n := range names {
		applied[n] = true
	}
	return applied, nil
}

// runMigrationInTx runs one migration and records it atomically.
func (s *PostgresStore) runMigrationInTx(ctx context.Context, migration migrations.Migration) error {
	return s.gdb.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m := &txMigrator{tx: tx}
		if err := migration.Up(ctx, m); err != nil {
			return err
		}
		return tx.Exec(
			`INSERT INTO go_schema_migrations (name) VALUES (?)`,
			migration.Name,
		).Error
	})
}
