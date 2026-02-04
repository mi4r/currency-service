package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_Generate(t *testing.T) {
	tests := []struct {
		name       string
		secret     string
		expiration time.Duration
		issuer     string
		userID     string
		username   string
		wantErr    bool
	}{
		{
			name:       "successful token generation",
			secret:     "test-secret-key-32-chars-long!!",
			expiration: 1 * time.Hour,
			issuer:     "test-issuer",
			userID:     "user-1",
			username:   "testuser",
			wantErr:    false,
		},
		{
			name:       "empty user id",
			secret:     "test-secret-key-32-chars-long!!",
			expiration: 1 * time.Hour,
			issuer:     "test-issuer",
			userID:     "",
			username:   "testuser",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager(tt.secret, tt.expiration, tt.issuer)

			token, expiresAt, err := manager.Generate(tt.userID, tt.username)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, token)
			assert.True(t, expiresAt.After(time.Now()))
		})
	}
}

func TestManager_Validate(t *testing.T) {
	tests := []struct {
		name           string
		setupManager   func() *Manager
		getToken       func(m *Manager) string
		expectedUserID string
		expectedUser   string
		expectedErr    error
	}{
		{
			name: "valid token",
			setupManager: func() *Manager {
				return NewManager("test-secret-key-32-chars-long!!", 1*time.Hour, "test-issuer")
			},
			getToken: func(m *Manager) string {
				token, _, _ := m.Generate("user-1", "testuser")
				return token
			},
			expectedUserID: "user-1",
			expectedUser:   "testuser",
			expectedErr:    nil,
		},
		{
			name: "invalid token format",
			setupManager: func() *Manager {
				return NewManager("test-secret-key-32-chars-long!!", 1*time.Hour, "test-issuer")
			},
			getToken: func(m *Manager) string {
				return "invalid-token"
			},
			expectedErr: ErrInvalidToken,
		},
		{
			name: "wrong secret",
			setupManager: func() *Manager {
				return NewManager("secret-key-2-32-chars-long!!!!!!", 1*time.Hour, "test-issuer")
			},
			getToken: func(m *Manager) string {
				otherManager := NewManager("secret-key-1-32-chars-long!!!!!!", 1*time.Hour, "test-issuer")
				token, _, _ := otherManager.Generate("user-1", "testuser")
				return token
			},
			expectedErr: ErrInvalidToken,
		},
		{
			name: "expired token",
			setupManager: func() *Manager {
				return NewManager("test-secret-key-32-chars-long!!", 1*time.Hour, "test-issuer")
			},
			getToken: func(m *Manager) string {
				expiredManager := NewManager("test-secret-key-32-chars-long!!", -1*time.Hour, "test-issuer")
				token, _, _ := expiredManager.Generate("user-1", "testuser")
				return token
			},
			expectedErr: ErrExpiredToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := tt.setupManager()
			token := tt.getToken(manager)

			claims, err := manager.Validate(token)

			if tt.expectedErr != nil {
				assert.Nil(t, claims)
				assert.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedUserID, claims.UserID)
			assert.Equal(t, tt.expectedUser, claims.Username)
		})
	}
}
