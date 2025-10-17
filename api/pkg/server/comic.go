package server

import (
	"fmt"

	"github.com/binocarlos/kai-stack/api/pkg/store"
	"github.com/binocarlos/kai-stack/api/pkg/types"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// ComicCreateRequest is the request body for creating a new comic
// ID, UserID, CreatedAt, and UpdatedAt are not included as they're set by the system
type ComicCreateRequest struct {
	Config *types.ComicConfig `json:"config"`
}

// ComicUpdateRequest is the request body for updating a comic
// ID, UserID, CreatedAt, and UpdatedAt are not included as they're managed by the system
type ComicUpdateRequest struct {
	Config *types.ComicConfig `json:"config"`
}

// ComicMapper handles mapping between Comic request DTOs and the Comic entity
type ComicMapper struct{}

// CreateToEntity converts a ComicCreateRequest to a Comic entity
func (m *ComicMapper) CreateToEntity(req *ComicCreateRequest) (*types.Comic, error) {
	if req.Config == nil {
		return nil, fmt.Errorf("config is required")
	}

	// Basic validation
	if req.Config.Name == "" {
		return nil, fmt.Errorf("comic name is required")
	}

	return &types.Comic{
		// ID and UserID will be set in BeforeCreate hook
		// CreatedAt and UpdatedAt are auto-managed by GORM
		Config: req.Config,
	}, nil
}

// UpdateToEntity applies a ComicUpdateRequest to an existing Comic entity
func (m *ComicMapper) UpdateToEntity(existing *types.Comic, req *ComicUpdateRequest) error {
	if req.Config == nil {
		return fmt.Errorf("config is required")
	}

	// Basic validation
	if req.Config.Name == "" {
		return fmt.Errorf("comic name is required")
	}

	// Update only the config field, preserving ID, UserID, and timestamps
	existing.Config = req.Config

	return nil
}

// ComicRouter provides CRUD operations for comics with custom logic
type ComicRouter struct {
	*ResourceRouter[types.Comic, ComicCreateRequest, ComicUpdateRequest]
	repo *store.ComicRepository
}

// NewComicRouter creates a new comic router with all necessary configuration
func NewComicRouter(apiServer *StackAPIServer, repo *store.ComicRepository) *ComicRouter {
	// Define hooks for custom comic behavior
	hooks := &ResourceHooks[types.Comic, ComicCreateRequest, ComicUpdateRequest]{
		BeforeCreate: func(c fiber.Ctx, comic *types.Comic) error {
			// Generate a new UUID for the comic
			comic.ID = uuid.New().String()

			// Set the UserID from the authenticated user context
			userID, ok := GetUserIDFromContext(c)
			if !ok || userID == "" {
				return fmt.Errorf("user ID is required to create a comic")
			}
			comic.UserID = userID

			return nil
		},
		AfterCreate: func(c fiber.Ctx, comic *types.Comic) error {
			// Could add logging, analytics, or trigger other services here
			return nil
		},
		BeforeUpdate: func(c fiber.Ctx, comic *types.Comic) error {
			// Could add validation or permission checks here
			// For example, ensure the user owns the comic they're updating
			userID, ok := GetUserIDFromContext(c)
			if !ok || userID == "" {
				return fmt.Errorf("user ID is required to update a comic")
			}

			if comic.UserID != userID {
				return fmt.Errorf("you can only update your own comics")
			}

			return nil
		},
	}

	// Configure the resource router
	config := &ResourceConfig[types.Comic, ComicCreateRequest, ComicUpdateRequest]{
		Hooks:      hooks,
		AuthConfig: DefaultAuthConfig(), // All operations require authentication
		Mapper:     &ComicMapper{},
	}

	resourceRouter := NewResourceRouter(apiServer, repo.Repository, config)

	return &ComicRouter{
		ResourceRouter: resourceRouter,
		repo:           repo,
	}
}

// RegisterRoutes registers all routes for the comic resource
func (cr *ComicRouter) RegisterRoutes(router fiber.Router) {
	// Register standard CRUD routes: GET, POST, GET /:id, PUT /:id, DELETE /:id
	cr.ResourceRouter.RegisterRoutes(router, "/comics")

	// Add custom route for getting comics by user
	// This demonstrates how to extend the base ResourceRouter with custom endpoints
	router.Get("/comics/user/:userId", cr.withAuth(cr.GetUserComics))
}

// GetUserComics returns all comics for a specific user
// This is a custom endpoint that uses the ComicRepository's LoadForUser method
func (cr *ComicRouter) GetUserComics(c fiber.Ctx) error {
	userID := c.Params("userId")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID parameter is required",
			"code":  "MISSING_USER_ID",
		})
	}

	// Optional: ensure users can only fetch their own comics
	authenticatedUserID, ok := GetUserIDFromContext(c)
	if !ok || authenticatedUserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You can only view your own comics",
			"code":  "FORBIDDEN",
		})
	}

	comics, err := cr.repo.LoadForUser(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load comics for user",
			"code":  "LOAD_USER_COMICS_FAILED",
		})
	}

	return c.Status(fiber.StatusOK).JSON(comics)
}
