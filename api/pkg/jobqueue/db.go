package jobqueue

import (
	"context"
	"fmt"

	"github.com/binocarlos/kai-stack/api/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// createPgxPool establishes a new pgx connection pool based on Store configuration.
func createPgxPool(ctx context.Context, dbConfig config.Database) (*pgxpool.Pool, error) {
	// Construct DSN (Data Source Name)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Database,
	)
	// Append SSL mode based on config
	if !dbConfig.SSL {
		dsn += "?sslmode=disable"
	} else {
		// Handle other SSL modes if necessary, e.g., require, verify-full
		// For now, assume sslmode=prefer or similar if dbConfig.SSL is true
		// and no specific mode is set in config. Adjust as needed.
	}

	// Add schema to search_path if specified
	if dbConfig.Schema != "" {
		// Check if we need '?' or '&' to append the search_path parameter
		separator := "?"
		if !dbConfig.SSL { // If sslmode=disable was added, use '&'
			separator = "&"
		} else if contains(dsn, "?") { // If other params exist (e.g. from SSL handling)
			separator = "&"
		}
		dsn += separator + "search_path=" + dbConfig.Schema
	}

	// Create pgx pool config from DSN
	pgxConf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config from DSN '%s': %w", dsn, err)
	}

	// Apply pool settings from config
	pgxConf.MaxConns = int32(dbConfig.MaxConns)
	// Use IdleConns for MinConns as a reasonable approximation if MinConns isn't explicit in config
	pgxConf.MinConns = int32(dbConfig.IdleConns)
	pgxConf.MaxConnLifetime = dbConfig.MaxConnLifetime
	pgxConf.MaxConnIdleTime = dbConfig.MaxConnIdleTime

	// Establish the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, pgxConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	// Ping the database to verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close() // Close pool if ping fails
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// Helper function to check if a string contains a specific character.
// Needed because standard library `strings.Contains` works on substrings.
func contains(s, char string) bool {
	for i := 0; i < len(s); i++ {
		if string(s[i]) == char {
			return true
		}
	}
	return false
}
