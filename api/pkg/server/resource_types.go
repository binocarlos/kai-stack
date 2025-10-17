package server

import "github.com/gofiber/fiber/v3"

// ResourceHooks defines optional lifecycle hooks for resource operations
// All hooks are optional (nil-safe) - they will only be called if provided
type ResourceHooks[T any, TCreate any, TUpdate any] struct {
	// BeforeCreate is called after parsing the request but before saving to database
	// Useful for validation, setting defaults, or populating fields from context
	BeforeCreate func(c fiber.Ctx, entity *T) error

	// AfterCreate is called after successfully saving to database
	// Useful for triggering side effects, logging, or returning custom responses
	AfterCreate func(c fiber.Ctx, entity *T) error

	// BeforeUpdate is called after parsing the request and fetching existing entity
	// The entity has been updated with new values from the request
	// Useful for validation or custom update logic
	BeforeUpdate func(c fiber.Ctx, entity *T) error

	// AfterUpdate is called after successfully updating in database
	AfterUpdate func(c fiber.Ctx, entity *T) error

	// BeforeDelete is called after fetching the entity but before deleting
	// Useful for validation or checking if deletion is allowed
	BeforeDelete func(c fiber.Ctx, id string, entity *T) error

	// AfterDelete is called after successfully deleting from database
	AfterDelete func(c fiber.Ctx, id string) error
}

// ResourceAuthConfig defines which operations require authentication
// By default, all operations require authentication
type ResourceAuthConfig struct {
	RequireAuthForList   bool
	RequireAuthForGet    bool
	RequireAuthForCreate bool
	RequireAuthForUpdate bool
	RequireAuthForDelete bool
}

// DefaultAuthConfig returns a config where all operations require authentication
func DefaultAuthConfig() ResourceAuthConfig {
	return ResourceAuthConfig{
		RequireAuthForList:   true,
		RequireAuthForGet:    true,
		RequireAuthForCreate: true,
		RequireAuthForUpdate: true,
		RequireAuthForDelete: true,
	}
}

// NoAuthConfig returns a config where no operations require authentication
func NoAuthConfig() ResourceAuthConfig {
	return ResourceAuthConfig{
		RequireAuthForList:   false,
		RequireAuthForGet:    false,
		RequireAuthForCreate: false,
		RequireAuthForUpdate: false,
		RequireAuthForDelete: false,
	}
}

// ResourceConfig bundles all configuration for a resource router
type ResourceConfig[T any, TCreate any, TUpdate any] struct {
	Hooks      *ResourceHooks[T, TCreate, TUpdate]
	AuthConfig ResourceAuthConfig
	Mapper     ResourceMapper[T, TCreate, TUpdate]
}

// DefaultResourceConfig returns a config with authentication enabled and default mapper
func DefaultResourceConfig[T any]() ResourceConfig[T, T, T] {
	return ResourceConfig[T, T, T]{
		Hooks:      nil,
		AuthConfig: DefaultAuthConfig(),
		Mapper:     NewDefaultMapper[T](),
	}
}
