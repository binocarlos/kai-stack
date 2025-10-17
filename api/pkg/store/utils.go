package store

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	_ "github.com/doug-martin/goqu/v9/dialect/postgres"        // postgres query builder
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres migrations
	_ "github.com/lib/pq"                                      // enable postgres driver

	"github.com/rs/zerolog/log"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type MigrationScript struct {
	Name   string `gorm:"primaryKey"`
	HasRun bool
}

type namedTable interface {
	TableName() string
}

// createFK creates a foreign key relationship between two tables.
//
// The argument `src` is the table with the field (`fk`) which refers to the field `pk` in the other (`dest`) table.
func createFK(db *gorm.DB, src, dst interface{}, fk, pk string, onDelete, onUpdate string) error {
	var (
		srcTableName string
		dstTableName string
	)

	sourceType := reflect.TypeOf(src)
	_, ok := sourceType.MethodByName("TableName")
	if ok {
		srcTableName = src.(namedTable).TableName()
	} else {
		srcTableName = db.NamingStrategy.TableName(sourceType.Name())
	}

	destinationType := reflect.TypeOf(dst)
	_, ok = destinationType.MethodByName("TableName")
	if ok {
		dstTableName = dst.(namedTable).TableName()
	} else {
		dstTableName = db.NamingStrategy.TableName(destinationType.Name())
	}

	// Dealing with custom table names that contain schema in them
	constraintName := "fk_" + strings.ReplaceAll(srcTableName, ".", "_") + "_" + strings.ReplaceAll(dstTableName, ".", "_")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if !db.Migrator().HasConstraint(src, constraintName) {
		err := db.WithContext(ctx).Exec(fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s) ON DELETE %s ON UPDATE %s",
			srcTableName,
			constraintName,
			fk,
			dstTableName,
			pk,
			onDelete,
			onUpdate)).Error
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	return nil
}

type Scanner interface {
	Scan(dest ...interface{}) error
}

// Available DB types
const (
	DatabaseTypePostgres = "postgres"
)

type connectConfig struct {
	host            string
	port            int
	schemaName      string
	database        string
	username        string
	password        string
	ssl             bool
	idleConns       int
	maxConns        int
	maxConnIdleTime time.Duration
	maxConnLifetime time.Duration
}

func connect(ctx context.Context, cfg connectConfig) (*gorm.DB, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("sql store startup deadline exceeded")
		default:
			log.Info().
				Str("host", cfg.host).
				Int("port", cfg.port).
				Str("database", cfg.database).
				Msg("connecting to DB")

			var (
				err       error
				dialector gorm.Dialector
			)

			// Read SSL setting from environment
			sslSettings := "sslmode=disable"
			if cfg.ssl {
				sslSettings = "sslmode=require"
			}

			dsn := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s %s",
				cfg.username, cfg.password, cfg.host, cfg.port, cfg.database, sslSettings)

			dialector = gormpostgres.Open(dsn)

			// gormConfig := &gorm.Config{
			// 	Logger: NewGormLogger(time.Second, true),
			// }

			gormConfig := &gorm.Config{}

			if cfg.schemaName != "" {
				gormConfig.NamingStrategy = schema.NamingStrategy{
					TablePrefix: cfg.schemaName + ".",
				}
			}

			db, err := gorm.Open(dialector, gormConfig)
			if err != nil {
				time.Sleep(1 * time.Second)

				log.Err(err).Msg("sql store connector can't reach DB, waiting")

				continue
			}

			db = db.Debug()

			sqlDB, err := db.DB()
			if err != nil {
				return nil, err
			}
			sqlDB.SetMaxIdleConns(cfg.idleConns)
			sqlDB.SetMaxOpenConns(cfg.maxConns)
			sqlDB.SetConnMaxIdleTime(cfg.maxConnIdleTime)
			sqlDB.SetConnMaxLifetime(cfg.maxConnLifetime)

			log.Info().Msg("sql store connected")

			// success
			return db, nil
		}
	}
}
