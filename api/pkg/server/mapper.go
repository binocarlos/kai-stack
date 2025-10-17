package server

// ResourceMapper defines how to convert request DTOs to entities
// TCreate: the request body type for creating a new entity (typically without ID)
// TUpdate: the request body type for updating an entity (typically without ID, which comes from URL)
// T: the actual entity type stored in the database
type ResourceMapper[T any, TCreate any, TUpdate any] interface {
	// CreateToEntity converts a create request DTO to a new entity
	CreateToEntity(create *TCreate) (*T, error)

	// UpdateToEntity applies update request DTO fields to an existing entity
	// This modifies the existing entity in place
	UpdateToEntity(existing *T, update *TUpdate) error
}

// DefaultMapper is a simple mapper that works when TCreate and TUpdate are the same type as T
// It uses type assertions and assumes the types are compatible
type DefaultMapper[T any] struct{}

// NewDefaultMapper creates a mapper that works when DTOs match the entity type
func NewDefaultMapper[T any]() *DefaultMapper[T] {
	return &DefaultMapper[T]{}
}

// CreateToEntity for DefaultMapper simply casts TCreate to T
// This works when TCreate is the same type as T
func (m *DefaultMapper[T]) CreateToEntity(create *T) (*T, error) {
	return create, nil
}

// UpdateToEntity for DefaultMapper replaces the existing entity with update data
// This works when TUpdate is the same type as T
// Note: This is a simple implementation that replaces all fields
// In practice, you might want to preserve certain fields like ID, CreatedAt, etc.
func (m *DefaultMapper[T]) UpdateToEntity(existing *T, update *T) error {
	*existing = *update
	return nil
}
