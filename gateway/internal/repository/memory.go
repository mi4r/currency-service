package repository

import (
	"context"
	"sync"

	"golang.org/x/crypto/bcrypt"

	"github.com/mi4r/currency-service/gateway/internal/domain"
)

// MemoryUserRepository implements an in-memory user repository with hardcoded users.
type MemoryUserRepository struct {
	users map[string]*domain.User
	mu    sync.RWMutex
}

// NewMemoryUserRepository creates a new in-memory user repository with default users.
func NewMemoryUserRepository() *MemoryUserRepository {
	repo := &MemoryUserRepository{
		users: make(map[string]*domain.User),
	}

	// Add hardcoded users
	repo.addUser("1", "admin", "admin123")
	repo.addUser("2", "user", "user123")

	return repo
}

func (r *MemoryUserRepository) addUser(id, username, password string) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	r.users[username] = &domain.User{
		ID:           id,
		Username:     username,
		PasswordHash: string(hash),
	}
}

// GetByUsername retrieves a user by username.
func (r *MemoryUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[username]
	if !ok {
		return nil, domain.ErrUserNotFound
	}

	return user, nil
}

// GetByID retrieves a user by ID.
func (r *MemoryUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.ID == id {
			return user, nil
		}
	}

	return nil, domain.ErrUserNotFound
}
