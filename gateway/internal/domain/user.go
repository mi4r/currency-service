package domain

import "errors"

// Domain errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// User represents a user in the system.
type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
}
