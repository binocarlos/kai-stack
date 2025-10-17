package jobqueue

import (
	"context"

	"github.com/binocarlos/kai-stack/api/pkg/config"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"
)

// this job is only run on the original video
// the transcoded video metadata is included in the transcoded video job
type TestArgs struct {
	Message string `json:"message"`
}

type TestResult struct {
	Args    TestArgs `json:"args"`
	JobID   int64    `json:"job_id"`
	Message string   `json:"message"`
}

func (TestArgs) Kind() string { return "test" }

type TestWorker struct {
	river.WorkerDefaults[TestArgs]
	config *config.Config
}

func newTestWorker(config *config.Config) *TestWorker {
	return &TestWorker{
		WorkerDefaults: river.WorkerDefaults[TestArgs]{},
		config:         config,
	}
}

// Work prints a friendly message then returns nil to mark the job completed.
func (w *TestWorker) Work(ctx context.Context, job *river.Job[TestArgs]) error {
	log.Info().Msgf("🟡 running test job: %d %+v", job.ID, job.Args)
	args := job.Args

	// Prepare the result with extracted metadata
	result := TestResult{
		Args:    args,
		JobID:   job.ID,
		Message: args.Message,
	}

	if err := postWorkerResult(*w.config, "/test", result); err != nil {
		log.Error().Msgf("error posting test job result: %s", err)
		return err
	}

	return nil
}
