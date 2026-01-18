package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mi4r/currency-service/pkg/jwt"
)

func TestAuth(t *testing.T) {
	secret := "test-secret-key-32-chars-long!!"

	tests := []struct {
		name           string
		setupAuth      func() string
		expectedStatus int
		checkContext   bool
		expectedUserID string
		expectedUser   string
	}{
		{
			name: "valid token",
			setupAuth: func() string {
				manager := jwt.NewManager(secret, 1*time.Hour, "test-issuer")
				token, _, _ := manager.Generate("user-1", "testuser")
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			checkContext:   true,
			expectedUserID: "user-1",
			expectedUser:   "testuser",
		},
		{
			name: "missing authorization header",
			setupAuth: func() string {
				return ""
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid format - no bearer prefix",
			setupAuth: func() string {
				return "InvalidFormat"
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid format - basic auth",
			setupAuth: func() string {
				return "Basic dXNlcjpwYXNz"
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid token",
			setupAuth: func() string {
				return "Bearer invalid-token"
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "expired token",
			setupAuth: func() string {
				manager := jwt.NewManager(secret, -1*time.Hour, "test-issuer")
				token, _, _ := manager.Generate("user-1", "testuser")
				return "Bearer " + token
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "wrong secret",
			setupAuth: func() string {
				wrongManager := jwt.NewManager("wrong-secret-key-32-chars-long!!", 1*time.Hour, "test-issuer")
				token, _, _ := wrongManager.Generate("user-1", "testuser")
				return "Bearer " + token
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwtManager := jwt.NewManager(secret, 1*time.Hour, "test-issuer")

			var capturedUserID, capturedUsername string
			handler := Auth(jwtManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedUserID = GetUserID(r.Context())
				capturedUsername = GetUsername(r.Context())
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			authHeader := tt.setupAuth()
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.checkContext {
				assert.Equal(t, tt.expectedUserID, capturedUserID)
				assert.Equal(t, tt.expectedUser, capturedUsername)
			}
		})
	}
}
