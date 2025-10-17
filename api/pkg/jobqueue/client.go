package jobqueue

import (
	"context"
	"fmt"

	"github.com/binocarlos/kai-stack/api/pkg/config"
	"github.com/binocarlos/kai-stack/api/pkg/store"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"
	"github.com/rs/zerolog/log"
)

// Client wraps a River client configured for insert-only behaviour
// so the rest of the API doesn't need to know about River internals.
// It exposes helper methods for enqueuing jobs.
// The Worker side is started by worker.go in a separate process.
type Client struct {
	ctx   context.Context
	river *river.Client[pgx.Tx]
	pool  *pgxpool.Pool
}

// NewClient constructs an insert-only River client using the provided
// database configuration details. It establishes a new connection pool using pgx.
// Note: we do NOT call Start because this component is only used for insertion.
func NewClient(ctx context.Context, config *config.Config, storeInstance *store.PostgresStore) (*Client, error) {
	// Create pgx pool using the shared function
	pool, err := createPgxPool(ctx, config.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool for client: %w", err)
	}

	// Create client struct first (river will be populated later)
	client := &Client{ctx: ctx, pool: pool}

	// Register workers, passing client as JobQueue interface
	workers := river.NewWorkers()
	river.AddWorker(workers, newTestWorker(config))

	// Create River client with pgxv5 driver
	// Note: River requires river.NewClient[pgx.Tx](...) for pgx with transaction support
	c, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		MaxAttempts:  config.Worker.MaxAttempts,
		ErrorHandler: newErrorHandler(config),
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: config.Worker.Concurrency},
		},
		Workers: workers,
	})
	if err != nil {
		pool.Close() // Close the pool if client creation fails
		return nil, fmt.Errorf("failed to create river client: %w", err)
	}

	// Populate the river field
	client.river = c

	return client, nil
}

func (c *Client) Start() error {
	return c.river.Start(c.ctx)
}

// Close cleans up resources, specifically the database pool connection.
func (c *Client) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}

func (c *Client) EnqueueTest(
	ctx context.Context,
	message string,
) (*rivertype.JobInsertResult, error) {
	args := TestArgs{
		Message: message,
	}
	return c.river.Insert(ctx, args, nil)
}

// EnqueueJob is a generic method to enqueue any River job
// The args parameter must implement river.JobArgs interface (have a Kind() method)
func (c *Client) EnqueueJob(ctx context.Context, args river.JobArgs) error {
	_, err := c.river.Insert(ctx, args, nil)
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}
	log.Info().Msgf("📋 Enqueued job: kind=%s", args.Kind())
	return nil
}
