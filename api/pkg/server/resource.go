package server

import (
	"fmt"

	"github.com/binocarlos/kai-stack/api/pkg/store"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

// ResourceRouter provides generic CRUD operations for a repository
// T: the entity type stored in the database
// TCreate: the request body type for creating entities
// TUpdate: the request body type for updating entities
type ResourceRouter[T any, TCreate any, TUpdate any] struct {
	repo      *store.Repository[T]
	config    ResourceConfig[T, TCreate, TUpdate]
	apiServer *StackAPIServer
}

// NewResourceRouter creates a new resource router with the given repository and configuration
// If config is nil, DefaultResourceConfig will be used (auth enabled, default mapper, no hooks)
func NewResourceRouter[T any, TCreate any, TUpdate any](
	apiServer *StackAPIServer,
	repo *store.Repository[T],
	config *ResourceConfig[T, TCreate, TUpdate],
) *ResourceRouter[T, TCreate, TUpdate] {
	if config == nil {
		defaultConfig := DefaultResourceConfig[T]()
		// Type assertion hack to convert from ResourceConfig[T,T,T] to ResourceConfig[T,TCreate,TUpdate]
		// This works because the config itself is generic
		config = &ResourceConfig[T, TCreate, TUpdate]{
			Hooks:      nil,
			AuthConfig: defaultConfig.AuthConfig,
			Mapper:     nil, // Will need to be provided if TCreate/TUpdate differ from T
		}
	}
	return &ResourceRouter[T, TCreate, TUpdate]{
		repo:      repo,
		config:    *config,
		apiServer: apiServer,
	}
}

// RegisterRoutes registers all CRUD routes for this resource
// path should be the base path for the resource (e.g., "/comics")
// This will create the following routes:
// - GET    {path}      -> List
// - POST   {path}      -> Create
// - GET    {path}/:id  -> Get
// - PUT    {path}/:id  -> Update
// - DELETE {path}/:id  -> Delete
func (rr *ResourceRouter[T, TCreate, TUpdate]) RegisterRoutes(router fiber.Router, path string) {
	// Apply auth middleware conditionally based on config
	listHandler := rr.List
	if rr.config.AuthConfig.RequireAuthForList {
		listHandler = rr.withAuth(listHandler)
	}

	createHandler := rr.Create
	if rr.config.AuthConfig.RequireAuthForCreate {
		createHandler = rr.withAuth(createHandler)
	}

	getHandler := rr.Get
	if rr.config.AuthConfig.RequireAuthForGet {
		getHandler = rr.withAuth(getHandler)
	}

	updateHandler := rr.Update
	if rr.config.AuthConfig.RequireAuthForUpdate {
		updateHandler = rr.withAuth(updateHandler)
	}

	deleteHandler := rr.Delete
	if rr.config.AuthConfig.RequireAuthForDelete {
		deleteHandler = rr.withAuth(deleteHandler)
	}

	router.Get(path, listHandler)
	router.Post(path, createHandler)
	router.Get(fmt.Sprintf("%s/:id", path), getHandler)
	router.Put(fmt.Sprintf("%s/:id", path), updateHandler)
	router.Delete(fmt.Sprintf("%s/:id", path), deleteHandler)
}

// withAuth wraps a handler with authentication middleware
func (rr *ResourceRouter[T, TCreate, TUpdate]) withAuth(handler fiber.Handler) fiber.Handler {
	return func(c fiber.Ctx) error {
		if err := rr.apiServer.RequireAuth(c); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
				"code":  "UNAUTHORIZED",
			})
		}
		return handler(c)
	}
}

// List returns all entities
func (rr *ResourceRouter[T, TCreate, TUpdate]) List(c fiber.Ctx) error {
	var entities []T
	if err := rr.repo.FindAll(&entities); err != nil {
		log.Error().Err(err).Msg("Failed to list entities")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve entities",
			"code":  "LIST_FAILED",
		})
	}

	return c.Status(fiber.StatusOK).JSON(entities)
}

// Get returns a single entity by ID
func (rr *ResourceRouter[T, TCreate, TUpdate]) Get(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID parameter is required",
			"code":  "MISSING_ID",
		})
	}

	var entity T
	if err := rr.repo.FindByID(id, &entity); err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to find entity")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Entity not found",
			"code":  "NOT_FOUND",
		})
	}

	return c.Status(fiber.StatusOK).JSON(entity)
}

// Create creates a new entity
func (rr *ResourceRouter[T, TCreate, TUpdate]) Create(c fiber.Ctx) error {
	// Parse request body
	req, err := getRequestData[TCreate](c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	// Map request to entity
	entity, err := rr.config.Mapper.CreateToEntity(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to map create request to entity")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request data",
			"code":  "MAPPING_FAILED",
		})
	}

	// Call BeforeCreate hook if provided
	if rr.config.Hooks != nil && rr.config.Hooks.BeforeCreate != nil {
		if err := rr.config.Hooks.BeforeCreate(c, entity); err != nil {
			log.Error().Err(err).Msg("BeforeCreate hook failed")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
				"code":  "BEFORE_CREATE_FAILED",
			})
		}
	}

	// Create entity in database
	if err := rr.repo.Create(entity); err != nil {
		log.Error().Err(err).Msg("Failed to create entity")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create entity",
			"code":  "CREATE_FAILED",
		})
	}

	// Call AfterCreate hook if provided
	if rr.config.Hooks != nil && rr.config.Hooks.AfterCreate != nil {
		if err := rr.config.Hooks.AfterCreate(c, entity); err != nil {
			log.Warn().Err(err).Msg("AfterCreate hook failed (entity already created)")
			// Don't return error since entity is already created
		}
	}

	return c.Status(fiber.StatusCreated).JSON(entity)
}

// Update updates an existing entity
func (rr *ResourceRouter[T, TCreate, TUpdate]) Update(c fiber.Ctx) error {
	// Get ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID parameter is required",
			"code":  "MISSING_ID",
		})
	}

	// Parse request body
	req, err := getRequestData[TUpdate](c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	// Fetch existing entity
	var entity T
	if err := rr.repo.FindByID(id, &entity); err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to find entity for update")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Entity not found",
			"code":  "NOT_FOUND",
		})
	}

	// Map update request to entity
	if err := rr.config.Mapper.UpdateToEntity(&entity, req); err != nil {
		log.Error().Err(err).Msg("Failed to map update request to entity")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request data",
			"code":  "MAPPING_FAILED",
		})
	}

	// Call BeforeUpdate hook if provided
	if rr.config.Hooks != nil && rr.config.Hooks.BeforeUpdate != nil {
		if err := rr.config.Hooks.BeforeUpdate(c, &entity); err != nil {
			log.Error().Err(err).Msg("BeforeUpdate hook failed")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
				"code":  "BEFORE_UPDATE_FAILED",
			})
		}
	}

	// Update entity in database
	if err := rr.repo.Update(&entity); err != nil {
		log.Error().Err(err).Msg("Failed to update entity")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update entity",
			"code":  "UPDATE_FAILED",
		})
	}

	// Call AfterUpdate hook if provided
	if rr.config.Hooks != nil && rr.config.Hooks.AfterUpdate != nil {
		if err := rr.config.Hooks.AfterUpdate(c, &entity); err != nil {
			log.Warn().Err(err).Msg("AfterUpdate hook failed (entity already updated)")
			// Don't return error since entity is already updated
		}
	}

	return c.Status(fiber.StatusOK).JSON(entity)
}

// Delete deletes an entity by ID
func (rr *ResourceRouter[T, TCreate, TUpdate]) Delete(c fiber.Ctx) error {
	// Get ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID parameter is required",
			"code":  "MISSING_ID",
		})
	}

	// Fetch existing entity for BeforeDelete hook
	var entity T
	fetchErr := rr.repo.FindByID(id, &entity)

	// Call BeforeDelete hook if provided (even if fetch failed, pass the id)
	if rr.config.Hooks != nil && rr.config.Hooks.BeforeDelete != nil {
		if err := rr.config.Hooks.BeforeDelete(c, id, &entity); err != nil {
			log.Error().Err(err).Msg("BeforeDelete hook failed")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
				"code":  "BEFORE_DELETE_FAILED",
			})
		}
	}

	// If we couldn't fetch the entity and no hook prevented deletion, return not found
	if fetchErr != nil {
		log.Error().Err(fetchErr).Str("id", id).Msg("Failed to find entity for deletion")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Entity not found",
			"code":  "NOT_FOUND",
		})
	}

	// Delete entity from database
	if err := rr.repo.Delete(id); err != nil {
		log.Error().Err(err).Msg("Failed to delete entity")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete entity",
			"code":  "DELETE_FAILED",
		})
	}

	// Call AfterDelete hook if provided
	if rr.config.Hooks != nil && rr.config.Hooks.AfterDelete != nil {
		if err := rr.config.Hooks.AfterDelete(c, id); err != nil {
			log.Warn().Err(err).Msg("AfterDelete hook failed (entity already deleted)")
			// Don't return error since entity is already deleted
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Entity deleted successfully",
		"id":      id,
	})
}
