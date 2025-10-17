package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/binocarlos/kai-stack/api/pkg/config"
	"github.com/binocarlos/kai-stack/api/pkg/types"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"        // postgres query builder
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres migrations
	_ "github.com/lib/pq"                                      // enable postgres driver

	"gorm.io/gorm"
)

type PostgresStore struct {
	cfg config.Database

	gdb *gorm.DB

	comics *ComicRepository
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
		cfg:    cfg,
		gdb:    gormDB,
		comics: NewComicRepository(gormDB),
	}

	if cfg.AutoMigrate {
		err = store.autoMigrate()
		if err != nil {
			return nil, fmt.Errorf("there was an error doing the automigration: %s", err.Error())
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

func (s *PostgresStore) autoMigrate() error {
	// If schema is specified, check if it exists and if not - create it
	if s.cfg.Schema != "" {
		err := s.gdb.WithContext(context.Background()).Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", s.cfg.Schema)).Error
		if err != nil {
			return err
		}
	}

	err := s.gdb.WithContext(context.Background()).AutoMigrate(
		&types.Comic{},
	)
	if err != nil {
		return err
	}

	// if err := createFK(s.gdb, types.Comic{}, types.User{}, "user_id", "id", "CASCADE", "CASCADE"); err != nil {
	// 	log.Err(err).Msg("failed to add DB FK")
	// }

	return nil
}
func (s *PostgresStore) SQLDB() (*sql.DB, error) {
	// expose the underlying *sql.DB for reuse in other subsystems (e.g. job queue)
	return s.gdb.DB()
}

// Comics returns the comic repository
func (s *PostgresStore) Comics() *ComicRepository {
	return s.comics
}
