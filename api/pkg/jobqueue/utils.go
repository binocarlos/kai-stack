package jobqueue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/binocarlos/kai-stack/api/pkg/config"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/rs/zerolog/log"
)

type JobErrorHandler struct {
	config *config.Config
}

type ErrorResult struct {
	JobID   int64  `json:"job_id"`
	VideoID string `json:"video_id"`
	Error   string `json:"error"`
}

type PanicResult struct {
	JobID   int64  `json:"job_id"`
	VideoID string `json:"video_id"`
	Trace   string `json:"trace"`
}

type LogResult struct {
	VideoID string   `json:"video_id"`
	Logs    []string `json:"logs"`
}

func newErrorHandler(config *config.Config) *JobErrorHandler {
	return &JobErrorHandler{
		config: config,
	}
}

func (errorHandler *JobErrorHandler) HandleError(ctx context.Context, job *rivertype.JobRow, err error) *river.ErrorHandlerResult {
	log.Error().
		Int64("job_id", job.ID).
		Str("kind", job.Kind).
		Int("attempts", job.Attempt).
		Int("max_attempts", job.MaxAttempts).
		Err(err).
		Msg("Job error occurred")

	if job.Attempt >= job.MaxAttempts {
		postError := postWorkerResult(*errorHandler.config, "/error", ErrorResult{
			JobID: job.ID,
			Error: err.Error(),
		})
		if postError != nil {
			log.Error().Msgf("error posting error job result: %s", postError)
			return &river.ErrorHandlerResult{
				SetCancelled: true,
			}
		}
	}
	return nil
}

// HandlePanic is called when a job panics
func (errorHandler *JobErrorHandler) HandlePanic(ctx context.Context, job *rivertype.JobRow, panicVal any, trace string) *river.ErrorHandlerResult {
	log.Error().
		Int64("job_id", job.ID).
		Str("kind", job.Kind).
		Interface("panic", panicVal).
		Str("stack", trace).
		Msg("Job panicked")

	if job.Attempt >= job.MaxAttempts {
		postError := postWorkerResult(*errorHandler.config, "/panic", PanicResult{
			JobID: job.ID,
			Trace: trace,
		})
		if postError != nil {
			log.Error().Msgf("error posting panic job result: %s", postError)
			return &river.ErrorHandlerResult{
				SetCancelled: true,
			}
		}
	}
	return nil
}

func postWorkerResult(config config.Config, apiPath string, body any) error {
	// Construct the full API endpoint URL
	baseURL, err := url.Parse(config.Worker.APIURL)
	if err != nil {
		return fmt.Errorf("invalid base URL in config: %w", err)
	}

	// Ensure BasePath starts with a slash and apiPath doesn't (or vice versa)
	// to avoid double slashes when joining.
	fullPath := strings.TrimSuffix(baseURL.Path, "/") + "/" + strings.TrimPrefix(config.WebServer.APIPath, "/")
	fullPath = strings.TrimSuffix(fullPath, "/") + strings.TrimPrefix(apiPath, "/")

	// Resolve the final path relative to the base URL
	apiURL := baseURL.ResolveReference(&url.URL{Path: fullPath})

	// Marshal the body into JSON
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", apiURL.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if config.Worker.Secret != "" {
		req.Header.Set("Authorization", "Bearer "+config.Worker.Secret)
	} else {
		// Potentially return an error if the secret is missing and required
		return fmt.Errorf("worker secret is missing in config, cannot authenticate")
	}

	// Send the request using the default HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request to %s: %w", apiURL.String(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if readErr != nil {
			return fmt.Errorf("worker request to %s failed with status %d", apiURL.String(), resp.StatusCode)
		}
		return fmt.Errorf("worker request to %s failed with status %d: %s", apiURL.String(), resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	// Exhaust body to allow connection reuse
	_, _ = io.Copy(io.Discard, resp.Body)

	return nil
}
