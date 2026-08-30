package models

import "time"

// Account represents the account database record.
type Account struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose in JSON responses
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RegisterRequest defines the input payload for POST /auth/register
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse defines the safe, non-sensitive output payload returned upon successful creation
type RegisterResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginRequest defines the input payload for POST /auth/login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SessionUser defines the non-sensitive user data stored in a session
type SessionUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// LoginResponse defines the successful response returned upon login
type LoginResponse struct {
	SessionToken string      `json:"session_token"`
	ExpiresIn    int         `json:"expires_in"` // in seconds
	User         SessionUser `json:"user"`
}
