package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/mi4r/currency-service/pkg/httputil"
	"github.com/mi4r/currency-service/pkg/jwt"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UsernameKey contextKey = "username"
)

// Auth creates a JWT authentication middleware.
func Auth(jwtManager *jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httputil.Unauthorized(w, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				httputil.Unauthorized(w, "invalid authorization header format")
				return
			}

			tokenString := parts[1]

			claims, err := jwtManager.Validate(tokenString)
			if err != nil {
				if err == jwt.ErrExpiredToken {
					httputil.Unauthorized(w, "token has expired")
					return
				}
				httputil.Unauthorized(w, "invalid token")
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID retrieves the user ID from context.
func GetUserID(ctx context.Context) string {
	if v := ctx.Value(UserIDKey); v != nil {
		return v.(string)
	}
	return ""
}

// GetUsername retrieves the username from context.
func GetUsername(ctx context.Context) string {
	if v := ctx.Value(UsernameKey); v != nil {
		return v.(string)
	}
	return ""
}
