package store

import (
	"context"
	"time"

	"github.com/binocarlos/kai-stack/api/pkg/types"
	pgvector "github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// ExampleRecordRepository is the skeleton's example repository: a thin wrapper
// over the generic Repository[T] (giving Create/FindByID/FindAll/Update/Delete)
// plus hand-written methods that show the access patterns these apps care about
// - loading by user, and vector similarity search over embeddings.
type ExampleRecordRepository struct {
	*Repository[types.ExampleRecord]
}

func NewExampleRecordRepository(db *gorm.DB) *ExampleRecordRepository {
	return &ExampleRecordRepository{
		Repository: NewRepository[types.ExampleRecord](db),
	}
}

// LoadForUser returns all example records belonging to a user. This is the
// canonical "override a method to add a custom query" example.
func (r *ExampleRecordRepository) LoadForUser(userID string) ([]types.ExampleRecord, error) {
	var records []types.ExampleRecord
	err := r.db.Where("user_id = ?", userID).Find(&records).Error
	return records, err
}

// UpsertEmbedding stores (or replaces) the embedding for a record. The vector
// is written through pgvector's type, which binds cleanly as a query argument
// so we never hand-format the "[1,2,3]" literal.
func (r *ExampleRecordRepository) UpsertEmbedding(ctx context.Context, recordID, content string, embedding []float32) error {
	now := time.Now().UnixMilli()
	return r.db.WithContext(ctx).Exec(`
INSERT INTO example_embeddings (record_id, content, embedding, created_at, updated_at)
VALUES (?, ?, ?::vector, ?, ?)
ON CONFLICT (record_id) DO UPDATE
SET content = EXCLUDED.content,
    embedding = EXCLUDED.embedding,
    updated_at = EXCLUDED.updated_at`,
		recordID, content, pgvector.NewVector(embedding), now, now,
	).Error
}

// SearchSimilar returns the records whose embeddings are closest to the query
// vector, ordered by cosine distance (smaller == more similar). The HNSW index
// created in migration 000002 makes this efficient.
func (r *ExampleRecordRepository) SearchSimilar(ctx context.Context, embedding []float32, limit int) ([]types.EmbeddingMatch, error) {
	if limit <= 0 {
		limit = 10
	}
	q := pgvector.NewVector(embedding)
	var matches []types.EmbeddingMatch
	err := r.db.WithContext(ctx).Raw(`
SELECT record_id, content, embedding <=> ?::vector AS distance
FROM example_embeddings
ORDER BY embedding <=> ?::vector
LIMIT ?`,
		q, q, limit,
	).Scan(&matches).Error
	return matches, err
}
