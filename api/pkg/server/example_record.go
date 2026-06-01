package server

import (
	"context"
	"fmt"

	"github.com/binocarlos/kai-stack/api/pkg/jobqueue"
	"github.com/binocarlos/kai-stack/api/pkg/store"
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

// ExampleRecordMapper maps between request DTOs and the ExampleRecord entity.
type ExampleRecordMapper struct{}

func (m *ExampleRecordMapper) CreateToEntity(req *ExampleRecordCreateRequest) (*types.ExampleRecord, error) {
	if req.Config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if req.Config.Name == "" {
		return nil, fmt.Errorf("example record name is required")
	}
	return &types.ExampleRecord{
		// ID and UserID are set in the BeforeCreate hook.
		Config: req.Config,
	}, nil
}

func (m *ExampleRecordMapper) UpdateToEntity(existing *types.ExampleRecord, req *ExampleRecordUpdateRequest) error {
	if req.Config == nil {
		return fmt.Errorf("config is required")
	}
	if req.Config.Name == "" {
		return fmt.Errorf("example record name is required")
	}
	existing.Config = req.Config
	return nil
}

// ExampleRecordRouter provides CRUD for example records plus custom endpoints.
type ExampleRecordRouter struct {
	*ResourceRouter[types.ExampleRecord, ExampleRecordCreateRequest, ExampleRecordUpdateRequest]
	repo     *store.ExampleRecordRepository
	jobqueue *jobqueue.Client
}

// NewExampleRecordRouter wires up the example record router. The AfterCreate
// hook enqueues a background job, demonstrating the enqueue -> worker round trip.
func NewExampleRecordRouter(apiServer *StackAPIServer, repo *store.ExampleRecordRepository) *ExampleRecordRouter {
	hooks := &ResourceHooks[types.ExampleRecord, ExampleRecordCreateRequest, ExampleRecordUpdateRequest]{
		BeforeCreate: func(c fiber.Ctx, record *types.ExampleRecord) error {
			record.ID = uuid.New().String()

			userID, ok := GetUserIDFromContext(c)
			if !ok || userID == "" {
				return fmt.Errorf("user ID is required to create an example record")
			}
			record.UserID = userID
			return nil
		},
		AfterCreate: func(c fiber.Ctx, record *types.ExampleRecord) error {
			// Detach background work from the request: enqueue a job that the
			// worker will pick up. Use a fresh context so the enqueue isn't
			// cancelled when the HTTP response is sent.
			if err := apiServer.jobqueue.Enqueue(context.Background(), jobqueue.ExampleJobKind, jobqueue.ExamplePayload{
				RecordID: record.ID,
				Message:  "example record created",
			}); err != nil {
				// Don't fail the request - the record is already created.
				log.Warn().Err(err).Str("record_id", record.ID).Msg("failed to enqueue example job")
			}
			return nil
		},
		BeforeUpdate: func(c fiber.Ctx, record *types.ExampleRecord) error {
			userID, ok := GetUserIDFromContext(c)
			if !ok || userID == "" {
				return fmt.Errorf("user ID is required to update an example record")
			}
			if record.UserID != userID {
				return fmt.Errorf("you can only update your own example records")
			}
			return nil
		},
	}

	config := &ResourceConfig[types.ExampleRecord, ExampleRecordCreateRequest, ExampleRecordUpdateRequest]{
		Hooks:      hooks,
		AuthConfig: DefaultAuthConfig(),
		Mapper:     &ExampleRecordMapper{},
	}

	resourceRouter := NewResourceRouter(apiServer, repo.Repository, config)

	return &ExampleRecordRouter{
		ResourceRouter: resourceRouter,
		repo:           repo,
		jobqueue:       apiServer.jobqueue,
	}
}

// RegisterRoutes registers CRUD routes plus the custom by-user endpoint.
func (rr *ExampleRecordRouter) RegisterRoutes(router fiber.Router) {
	rr.ResourceRouter.RegisterRoutes(router, "/example-records")

	// Custom route showing how to extend the base router with the repository's
	// hand-written methods.
	router.Get("/example-records/user/:userId", rr.withAuth(rr.GetUserRecords))
}

// GetUserRecords returns all example records for a user, via LoadForUser.
func (rr *ExampleRecordRouter) GetUserRecords(c fiber.Ctx) error {
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

	records, err := rr.repo.LoadForUser(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load example records for user",
			"code":  "LOAD_USER_RECORDS_FAILED",
		})
	}

	return c.Status(fiber.StatusOK).JSON(records)
}
