package types

// ExampleConfig is the JSONB payload stored against an ExampleRecord.
// It stands in for whatever domain config a real record would carry.
type ExampleConfig struct {
	Name        string `json:"name" gorm:"type:varchar(255);not null"`
	Description string `json:"description" gorm:"type:text"`
}

// ExampleRecord is the skeleton's example table: a single relational row
// per user. It is intentionally thin - the interesting access patterns
// (vector search) live alongside it in store/example_record.go.
type ExampleRecord struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID    string         `json:"user_id" gorm:"column:user_id;type:varchar(36);not null"`
	CreatedAt int64          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt int64          `json:"updated_at" gorm:"autoUpdateTime"`
	Config    *ExampleConfig `json:"config" gorm:"type:jsonb"`
}

// EmbeddingMatch is a single result from a vector similarity search:
// the embedding row plus its cosine distance to the query vector
// (smaller distance == more similar).
type EmbeddingMatch struct {
	RecordID string  `json:"record_id"`
	Content  string  `json:"content"`
	Distance float64 `json:"distance"`
}

// Job is one row in the SKIP LOCKED background-job queue. The queue is a
// plain Postgres table (see migrations/000003_jobs.go) polled by the worker -
// no external dependencies. Rows are managed via raw SQL in store/job.go, so
// this struct is just the shape a claimed job scans into.
type Job struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Payload     []byte `json:"payload"`
	Status      string `json:"status"`
	Attempts    int    `json:"attempts"`
	MaxAttempts int    `json:"max_attempts"`
	LastError   string `json:"last_error"`
}
