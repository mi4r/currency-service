package service

import (
	"context"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/mi4r/currency-service/gateway/internal/domain"
	"github.com/mi4r/currency-service/pkg/jwt"
)

// AuthResult represents the result of a successful authentication.
type AuthResult struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AuthService handles authentication operations.
type AuthService struct {
	userRepo   UserRepository
	jwtManager *jwt.Manager
	logger     *slog.Logger
}

// NewAuthService creates a new auth service.
func NewAuthService(userRepo UserRepository, jwtManager *jwt.Manager, logger *slog.Logger) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// Login authenticates a user and returns a JWT token.
func (s *AuthService) Login(ctx context.Context, username, password string) (*AuthResult, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		s.logger.Warn("login failed: user not found",
			slog.String("username", username),
		)
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.Warn("login failed: invalid password",
			slog.String("username", username),
		)
		return nil, domain.ErrInvalidCredentials
	}

	token, expiresAt, err := s.jwtManager.Generate(user.ID, user.Username)
	if err != nil {
		s.logger.Error("failed to generate token",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("user logged in successfully",
		slog.String("username", username),
	)

	return &AuthResult{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}
