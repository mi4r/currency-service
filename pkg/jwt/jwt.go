package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Claims represents the JWT claims structure.
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Manager handles JWT token operations.
type Manager struct {
	secret     []byte
	expiration time.Duration
	issuer     string
}

// NewManager creates a new JWT manager.
func NewManager(secret string, expiration time.Duration, issuer string) *Manager {
	return &Manager{
		secret:     []byte(secret),
		expiration: expiration,
		issuer:     issuer,
	}
}

// Generate creates a new JWT token for the given user.
func (m *Manager) Generate(userID, username string) (string, time.Time, error) {
	expiresAt := time.Now().Add(m.expiration)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    m.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// Validate parses and validates the JWT token.
func (m *Manager) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
