package migrations

import "context"

// VectorSupport enables pgvector and creates the example_embeddings table: one
// embedding per example_record, with an HNSW index for cosine similarity
// search. This is the core capability the skeleton exists to demonstrate.
//
// The dimension below (1536) matches OpenAI's text-embedding-3-small. Change
// it to match your embedding model - note that altering the column dimension
// later requires a new migration that rebuilds the column and index.
func VectorSupport(ctx context.Context, m Migrator) error {
	return m.ExecSQL(ctx, `
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS example_embeddings (
    record_id  varchar(36) PRIMARY KEY REFERENCES example_records(id) ON DELETE CASCADE,
    content    text NOT NULL,
    embedding  vector(1536),
    created_at bigint NOT NULL,
    updated_at bigint NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_example_embeddings_hnsw
    ON example_embeddings USING hnsw (embedding vector_cosine_ops);
`)
}
