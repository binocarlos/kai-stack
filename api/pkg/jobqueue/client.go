package jobqueue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/binocarlos/kai-stack/api/pkg/config"
	"github.com/binocarlos/kai-stack/api/pkg/store"
	"github.com/rs/zerolog/log"
)

// Handler processes a single job's payload. Returning an error triggers a
// retry (up to the job's max_attempts); returning nil marks it completed.
type Handler func(ctx context.Context, payload json.RawMessage) error

// Client is the public face of the background-job queue. It is shared by the
// API (which enqueues jobs) and the worker (which runs them via Start). The
// queue itself is just a Postgres table - see store.JobRepository.
type Client struct {
	cfg      *config.Config
	store    *store.PostgresStore
	handlers map[string]Handler
}

// NewClient builds a queue client and registers the built-in example handler.
// Register additional handlers with Register before calling Start.
func NewClient(cfg *config.Config, st *store.PostgresStore) *Client {
	c := &Client{
		cfg:      cfg,
		store:    st,
		handlers: map[string]Handler{},
	}
	registerExampleHandler(c)
	return c
}

// Register associates a handler with a job kind. Not safe to call concurrently
// with Start; register everything up front.
func (c *Client) Register(kind string, handler Handler) {
	c.handlers[kind] = handler
}

// Enqueue persists a job to be picked up by a worker. It refuses kinds with no
// registered handler so typos surface immediately rather than as stuck jobs.
func (c *Client) Enqueue(ctx context.Context, kind string, payload any) error {
	if _, ok := c.handlers[kind]; !ok {
		return fmt.Errorf("no handler registered for job kind %q", kind)
	}
	if _, err := c.store.Jobs().Enqueue(ctx, kind, payload); err != nil {
		return err
	}
	log.Info().Str("kind", kind).Msg("📋 enqueued job")
	return nil
}
