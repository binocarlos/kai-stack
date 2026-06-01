package server

import (
	"context"
	"fmt"

	"github.com/binocarlos/kai-stack/api/pkg/jobqueue"
	"github.com/binocarlos/kai-stack/api/pkg/types"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// ExampleRecordCreateRequest is the request body for creating an example record.
// ID, UserID, CreatedAt, and UpdatedAt are set by the system.
type ExampleRecordCreateRequest struct {
	Config *types.ExampleConfig `json:"config"`
}

// ExampleRecordUpdateRequest is the request body for updating an example record.
type ExampleRecordUpdateRequest struct {
	Config *types.ExampleConfig `json:"config"`
}

// RegisterExampleRecordRoutes wires up CRUD for example records plus a custom
// by-user endpoint. This is the canonical "how to add a resource" example:
// plain handler methods on StackAPIServer that talk to the store directly,
// mirroring user.go. All routes require auth.
func (apiServer *StackAPIServer) RegisterExampleRecordRoutes() {
	records := apiServer.router.Group("/example-records", apiServer.RequireAuth)
	records.Get("/", apiServer.ListExampleRecords)
	records.Post("/", apiServer.CreateExampleRecord)
	records.Get("/:id", apiServer.GetExampleRecord)
	records.Put("/:id", apiServer.UpdateExampleRecord)
	records.Delete("/:id", apiServer.DeleteExampleRecord)

	// Custom route showing how to extend CRUD with a hand-written repository
	// method (LoadForUser).
	records.Get("/user/:userId", apiServer.ListUserExampleRecords)
}

// validateExampleConfig enforces the same rules the old mapper did.
func validateExampleConfig(config *types.ExampleConfig) error {
	if config == nil {
		return fmt.Errorf("config is required")
	}
	if config.Name == "" {
		return fmt.Errorf("example record name is required")
	}
	return nil
}

// ListExampleRecords returns all example records.
func (apiServer *StackAPIServer) ListExampleRecords(c fiber.Ctx) error {
	var records []types.ExampleRecord
	if err := apiServer.store.ExampleRecords().FindAll(&records); err != nil {
		log.Error().Err(err).Msg("Failed to list example records")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve example records",
			"code":  "LIST_FAILED",
		})
	}
	return c.Status(fiber.StatusOK).JSON(records)
}

// GetExampleRecord returns a single example record by ID.
func (apiServer *StackAPIServer) GetExampleRecord(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID parameter is required",
			"code":  "MISSING_ID",
		})
	}

	var record types.ExampleRecord
	if err := apiServer.store.ExampleRecords().FindByID(id, &record); err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to find example record")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Example record not found",
			"code":  "NOT_FOUND",
		})
	}
	return c.Status(fiber.StatusOK).JSON(record)
}

// CreateExampleRecord creates an example record owned by the authenticated user,
// then enqueues a background job. The enqueue uses a fresh context so it isn't
// cancelled when the HTTP response is sent, and a failure to enqueue does not
// fail the request (the record is already created).
func (apiServer *StackAPIServer) CreateExampleRecord(c fiber.Ctx) error {
	req, err := getRequestData[ExampleRecordCreateRequest](c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	if err := validateExampleConfig(req.Config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
			"code":  "INVALID_REQUEST",
		})
	}

	userID, ok := GetUserIDFromContext(c)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user ID is required to create an example record",
			"code":  "UNAUTHORIZED",
		})
	}

	record := &types.ExampleRecord{
		ID:     uuid.New().String(),
		UserID: userID,
		Config: req.Config,
	}

	if err := apiServer.store.ExampleRecords().Create(record); err != nil {
		log.Error().Err(err).Msg("Failed to create example record")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create example record",
			"code":  "CREATE_FAILED",
		})
	}

	if err := apiServer.jobqueue.Enqueue(context.Background(), jobqueue.ExampleJobKind, jobqueue.ExamplePayload{
		RecordID: record.ID,
		Message:  "example record created",
	}); err != nil {
		log.Warn().Err(err).Str("record_id", record.ID).Msg("failed to enqueue example job")
	}

	return c.Status(fiber.StatusCreated).JSON(record)
}

// UpdateExampleRecord updates an example record the authenticated user owns.
func (apiServer *StackAPIServer) UpdateExampleRecord(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID parameter is required",
			"code":  "MISSING_ID",
		})
	}

	req, err := getRequestData[ExampleRecordUpdateRequest](c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	if err := validateExampleConfig(req.Config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
			"code":  "INVALID_REQUEST",
		})
	}

	var record types.ExampleRecord
	if err := apiServer.store.ExampleRecords().FindByID(id, &record); err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to find example record for update")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Example record not found",
			"code":  "NOT_FOUND",
		})
	}

	userID, ok := GetUserIDFromContext(c)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user ID is required to update an example record",
			"code":  "UNAUTHORIZED",
		})
	}
	if record.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "you can only update your own example records",
			"code":  "FORBIDDEN",
		})
	}

	record.Config = req.Config

	if err := apiServer.store.ExampleRecords().Update(&record); err != nil {
		log.Error().Err(err).Msg("Failed to update example record")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update example record",
			"code":  "UPDATE_FAILED",
		})
	}

	return c.Status(fiber.StatusOK).JSON(record)
}

// DeleteExampleRecord deletes an example record by ID.
func (apiServer *StackAPIServer) DeleteExampleRecord(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID parameter is required",
			"code":  "MISSING_ID",
		})
	}

	var record types.ExampleRecord
	if err := apiServer.store.ExampleRecords().FindByID(id, &record); err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to find example record for deletion")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Example record not found",
			"code":  "NOT_FOUND",
		})
	}

	if err := apiServer.store.ExampleRecords().Delete(id); err != nil {
		log.Error().Err(err).Msg("Failed to delete example record")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete example record",
			"code":  "DELETE_FAILED",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Example record deleted successfully",
		"id":      id,
	})
}

// ListUserExampleRecords returns all example records for a user, via the
// hand-written LoadForUser query. Users may only view their own records.
func (apiServer *StackAPIServer) ListUserExampleRecords(c fiber.Ctx) error {
	userID := c.Params("userId")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID parameter is required",
			"code":  "MISSING_USER_ID",
		})
	}

	authenticatedUserID, ok := GetUserIDFromContext(c)
	if !ok || authenticatedUserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You can only view your own example records",
			"code":  "FORBIDDEN",
		})
	}

	records, err := apiServer.store.ExampleRecords().LoadForUser(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load example records for user",
			"code":  "LOAD_USER_RECORDS_FAILED",
		})
	}

	return c.Status(fiber.StatusOK).JSON(records)
}
