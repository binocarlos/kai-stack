package store

import (
	"github.com/binocarlos/kai-stack/api/pkg/types"
	"gorm.io/gorm"
)

type ComicRepository struct {
	*Repository[types.Comic]
}

func NewComicRepository(db *gorm.DB) *ComicRepository {
	return &ComicRepository{
		Repository: NewRepository[types.Comic](db),
	}
}

func (r *ComicRepository) LoadForUser(userID string) ([]types.Comic, error) {
	var comics []types.Comic
	err := r.db.Where("user = ?", userID).Find(&comics).Error
	return comics, err
}
