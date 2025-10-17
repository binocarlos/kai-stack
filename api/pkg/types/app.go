package types

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
}
