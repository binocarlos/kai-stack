package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Roles is a list of authorization roles persisted as a JSONB column. It
// implements sql.Scanner / driver.Valuer so it round-trips cleanly through both
// GORM and raw queries.
type Roles []string

func (r *Roles) Scan(value any) error {
	if value == nil {
		*r = Roles{}
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("unsupported type for Roles: %T", value)
	}
	return json.Unmarshal(data, r)
}

func (r Roles) Value() (driver.Value, error) {
	if r == nil {
		return "[]", nil
	}
	b, err := json.Marshal(r)
	return string(b), err
}

// Profile is the app's canonical user. ID is our own UUID; (AuthProvider,
// AuthSubject) maps it to whichever auth provider verified the request.
type Profile struct {
	ID           string    `json:"id" gorm:"column:id"`
	AuthProvider string    `json:"auth_provider" gorm:"column:auth_provider"`
	AuthSubject  string    `json:"auth_subject" gorm:"column:auth_subject"`
	Email        string    `json:"email" gorm:"column:email"`
	Roles        Roles     `json:"roles" gorm:"column:roles"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (Profile) TableName() string { return "profiles" }

type User struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

// LoginRequest represents the expected request body for login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the response body for successful login
type LoginResponse struct {
	Token string `json:"token"`
}

type UserStatusResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Roles  Roles  `json:"roles"`
}
