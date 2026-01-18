package service

import (
	"context"

	"github.com/mi4r/currency-service/gateway/internal/domain"
)

// UserRepository defines the interface for user data operations.
// This interface is defined in the service layer (consumer) following Dependency Inversion Principle.
type UserRepository interface {
	// GetByUsername retrieves a user by username.
	GetByUsername(ctx context.Context, username string) (*domain.User, error)

	// GetByID retrieves a user by ID.
	GetByID(ctx context.Context, id string) (*domain.User, error)
}
