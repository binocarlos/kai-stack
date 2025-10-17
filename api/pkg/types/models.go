package types

type ComicConfig struct {
	Name        string `json:"name" gorm:"type:varchar(255);not null"`
	Description string `json:"description" gorm:"type:text"`
}

type Comic struct {
	ID        string       `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID    string       `json:"user_id" gorm:"type:varchar(36);not null"`
	CreatedAt int64        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt int64        `json:"updated_at" gorm:"autoUpdateTime"`
	Config    *ComicConfig `json:"config" gorm:"type:jsonb"`
}
