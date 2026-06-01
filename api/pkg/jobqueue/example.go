package jobqueue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

// ExampleJobKind is the kind for the skeleton's demonstration job. The API
// enqueues one of these after creating an example record (see
// CreateExampleRecord in the server package); the worker processes it in the
// background.
const ExampleJobKind = "example"

// ExamplePayload is the JSON payload for an example job.
type ExamplePayload struct {
	RecordID string `json:"record_id"`
	Message  string `json:"message"`
}

// registerExampleHandler wires up the built-in example handler. Replace or
// remove this when building a real app - it exists to show the enqueue ->
// process round trip end to end.
func registerExampleHandler(c *Client) {
	c.Register(ExampleJobKind, func(ctx context.Context, payload json.RawMessage) error {
		var args ExamplePayload
		if err := json.Unmarshal(payload, &args); err != nil {
			return fmt.Errorf("invalid example payload: %w", err)
		}
		log.Info().
			Str("record_id", args.RecordID).
			Str("message", args.Message).
			Msg("🟢 processed example job")
		return nil
	})
}
