package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/mi4r/currency-service/gateway/internal/domain"
)

func TestMemoryUserRepository_GetByUsername(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		expectUser  bool
		expectedID  string
		expectedErr error
	}{
		{
			name:       "existing admin user",
			username:   "admin",
			expectUser: true,
			expectedID: "1",
		},
		{
			name:       "existing regular user",
			username:   "user",
			expectUser: true,
			expectedID: "2",
		},
		{
			name:        "non-existing user",
			username:    "nonexistent",
			expectUser:  false,
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name:        "empty username",
			username:    "",
			expectUser:  false,
			expectedErr: domain.ErrUserNotFound,
		},
	}

	repo := NewMemoryUserRepository()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetByUsername(context.Background(), tt.username)

			if tt.expectedErr != nil {
				assert.Nil(t, user)
				assert.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedID, user.ID)
			assert.Equal(t, tt.username, user.Username)
			assert.NotEmpty(t, user.PasswordHash)
		})
	}
}

func TestMemoryUserRepository_GetByID(t *testing.T) {
	tests := []struct {
		name             string
		id               string
		expectUser       bool
		expectedUsername string
		expectedErr      error
	}{
		{
			name:             "existing user id 1",
			id:               "1",
			expectUser:       true,
			expectedUsername: "admin",
		},
		{
			name:             "existing user id 2",
			id:               "2",
			expectUser:       true,
			expectedUsername: "user",
		},
		{
			name:        "non-existing user id",
			id:          "999",
			expectUser:  false,
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name:        "empty id",
			id:          "",
			expectUser:  false,
			expectedErr: domain.ErrUserNotFound,
		},
	}

	repo := NewMemoryUserRepository()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetByID(context.Background(), tt.id)

			if tt.expectedErr != nil {
				assert.Nil(t, user)
				assert.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.id, user.ID)
			assert.Equal(t, tt.expectedUsername, user.Username)
		})
	}
}

func TestMemoryUserRepository_PasswordHashing(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		plainPassword  string
		shouldValidate bool
	}{
		{
			name:           "admin password valid",
			username:       "admin",
			plainPassword:  "admin123",
			shouldValidate: true,
		},
		{
			name:           "user password valid",
			username:       "user",
			plainPassword:  "user123",
			shouldValidate: true,
		},
		{
			name:           "admin password invalid",
			username:       "admin",
			plainPassword:  "wrongpassword",
			shouldValidate: false,
		},
	}

	repo := NewMemoryUserRepository()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetByUsername(context.Background(), tt.username)
			require.NoError(t, err)

			err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(tt.plainPassword))

			if tt.shouldValidate {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
