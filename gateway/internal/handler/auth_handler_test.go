package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mi4r/currency-service/gateway/internal/domain"
	"github.com/mi4r/currency-service/gateway/internal/service"
	"github.com/mi4r/currency-service/pkg/logger"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (*service.AuthResult, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AuthResult), args.Error(1)
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService)
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name: "successful login",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Login", mock.Anything, "testuser", "password123").Return(&service.AuthResult{
					Token:     "jwt-token-here",
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "jwt-token-here", body["token"])
				assert.NotEmpty(t, body["expires_at"])
			},
		},
		{
			name:           "invalid json body",
			requestBody:    "invalid json",
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "invalid request body", body["error"])
			},
		},
		{
			name: "missing username",
			requestBody: map[string]string{
				"password": "password123",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "username and password are required", body["error"])
			},
		},
		{
			name: "missing password",
			requestBody: map[string]string{
				"username": "testuser",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "username and password are required", body["error"])
			},
		},
		{
			name: "empty username",
			requestBody: map[string]string{
				"username": "",
				"password": "password123",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "username and password are required", body["error"])
			},
		},
		{
			name: "empty password",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "username and password are required", body["error"])
			},
		},
		{
			name: "invalid credentials",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "wrongpassword",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Login", mock.Anything, "testuser", "wrongpassword").Return(nil, domain.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "invalid credentials", body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockAuthService)
			tt.setupMock(mockSvc)
			h := NewAuthHandler(mockSvc, logger.New("error"))

			var bodyBytes []byte
			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.Login(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var body map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &body)
			require.NoError(t, err)

			tt.checkResponse(t, body)
			mockSvc.AssertExpectations(t)
		})
	}
}
