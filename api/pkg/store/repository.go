package store

import (
	"gorm.io/gorm"
)

type Repository[T any] struct {
	db *gorm.DB
}

func NewRepository[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{db: db}
}

func (r *Repository[T]) Create(entity *T) error {
	return r.db.Create(entity).Error
}

func (r *Repository[T]) FindByID(id string, entity *T) error {
	return r.db.First(entity, "id = ?", id).Error
}

func (r *Repository[T]) FindAll(entities *[]T) error {
	return r.db.Find(entities).Error
}

func (r *Repository[T]) Update(entity *T) error {
	return r.db.Save(entity).Error
}

func (r *Repository[T]) Delete(id string) error {
	var entity T
	return r.db.Delete(&entity, "id = ?", id).Error
}

func (r *Repository[T]) Where(query interface{}, args ...interface{}) *gorm.DB {
	return r.db.Where(query, args...)
}

// Usage:
// configRepo := repository.New[Config](db)
// err := configRepo.Create(&config)
// err := configRepo.FindByID("123", &config)
