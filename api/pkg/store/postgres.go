package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/binocarlos/kai-stack/api/pkg/config"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres" // postgres query builder
	_ "github.com/lib/pq"                               // enable postgres driver

	"gorm.io/gorm"
)

type PostgresStore struct {
	cfg config.Database

	gdb *gorm.DB

	exampleRecords *ExampleRecordRepository
	jobs           *JobRepository
	profiles       *ProfileRepository
}

func NewPostgresStore(
	cfg config.Database,
) (*PostgresStore, error) {

	// Waiting for connection
	gormDB, err := connect(context.Background(), connectConfig{
		host:            cfg.Host,
		port:            cfg.Port,
		schemaName:      cfg.Schema,
		database:        cfg.Database,
		username:        cfg.Username,
		password:        cfg.Password,
		ssl:             cfg.SSL,
		idleConns:       cfg.IdleConns,
		maxConns:        cfg.MaxConns,
		maxConnIdleTime: cfg.MaxConnIdleTime,
		maxConnLifetime: cfg.MaxConnLifetime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Postgres: %w", err)
	}

	store := &PostgresStore{
		cfg:            cfg,
		gdb:            gormDB,
		exampleRecords: NewExampleRecordRepository(gormDB),
		jobs:           NewJobRepository(gormDB),
		profiles:       NewProfileRepository(gormDB),
	}

	// Schema is owned by our own migration system (see migrations.go), not by
	// GORM AutoMigrate. AutoMigrate=true means "apply pending migrations on boot".
	if cfg.AutoMigrate {
		if cfg.Schema != "" {
			if err := gormDB.WithContext(context.Background()).
				Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", cfg.Schema)).Error; err != nil {
				return nil, fmt.Errorf("failed to create schema %s: %w", cfg.Schema, err)
			}
		}
		if err := store.runMigrations(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return store, nil
}

func (s *PostgresStore) Close() error {
	sqlDB, err := s.gdb.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// SQLDB exposes the underlying *sql.DB for reuse in other subsystems.
func (s *PostgresStore) SQLDB() (*sql.DB, error) {
	return s.gdb.DB()
}

// ExampleRecords returns the example_record repository (relational + vector).
func (s *PostgresStore) ExampleRecords() *ExampleRecordRepository {
	return s.exampleRecords
}

// Jobs returns the background-job queue repository.
func (s *PostgresStore) Jobs() *JobRepository {
	return s.jobs
}

// Profiles returns the canonical-user repository.
func (s *PostgresStore) Profiles() *ProfileRepository {
	return s.profiles
}
