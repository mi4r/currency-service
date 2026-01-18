package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/mi4r/currency-service/gateway/internal/domain"
	"github.com/mi4r/currency-service/pkg/httputil"
)

// AuthHandler handles authentication requests.
type AuthHandler struct {
	authService AuthService
	logger      *slog.Logger
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authService AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		httputil.BadRequest(w, "username and password are required")
		return
	}

	result, err := h.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			httputil.Unauthorized(w, "invalid credentials")
			return
		}
		h.logger.Error("login error", slog.String("error", err.Error()))
		httputil.InternalError(w, "internal server error")
		return
	}

	httputil.Success(w, LoginResponse{
		Token:     result.Token,
		ExpiresAt: result.ExpiresAt,
	})
}
