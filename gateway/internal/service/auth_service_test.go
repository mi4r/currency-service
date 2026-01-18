package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/mi4r/currency-service/gateway/internal/domain"
	"github.com/mi4r/currency-service/pkg/jwt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func hashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		password    string
		setupMock   func(*MockUserRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "successful login",
			username: "testuser",
			password: "password123",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByUsername", mock.Anything, "testuser").Return(&domain.User{
					ID:           "1",
					Username:     "testuser",
					PasswordHash: hashPassword("password123"),
				}, nil)
			},
			wantErr: false,
		},
		{
			name:     "user not found",
			username: "nonexistent",
			password: "password123",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByUsername", mock.Anything, "nonexistent").Return(nil, domain.ErrUserNotFound)
			},
			wantErr:     true,
			expectedErr: domain.ErrInvalidCredentials,
		},
		{
			name:     "invalid password",
			username: "testuser",
			password: "wrongpassword",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByUsername", mock.Anything, "testuser").Return(&domain.User{
					ID:           "1",
					Username:     "testuser",
					PasswordHash: hashPassword("password123"),
				}, nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrInvalidCredentials,
		},
		{
			name:     "empty password",
			username: "testuser",
			password: "",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByUsername", mock.Anything, "testuser").Return(&domain.User{
					ID:           "1",
					Username:     "testuser",
					PasswordHash: hashPassword("password123"),
				}, nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)
			jwtManager := jwt.NewManager("test-secret-key-32-chars-long!!", 1*time.Hour, "test-issuer")
			svc := NewAuthService(mockRepo, jwtManager, newTestLogger())

			result, err := svc.Login(context.Background(), tt.username, tt.password)

			if tt.wantErr {
				assert.Nil(t, result)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, result.Token)
				assert.True(t, result.ExpiresAt.After(time.Now()))
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
