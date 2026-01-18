package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mi4r/currency-service/gateway/internal/middleware"
	"github.com/mi4r/currency-service/pkg/httputil"
	"github.com/mi4r/currency-service/pkg/jwt"
)

// RegisterRoutes registers all routes for the gateway service.
func RegisterRoutes(
	r chi.Router,
	authHandler *AuthHandler,
	currencyHandler *CurrencyHandler,
	jwtManager *jwt.Manager,
	logger *slog.Logger,
) {
	// Global middleware
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logging(logger))

	// Health check (public)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httputil.Success(w, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			// r.Use(middleware.Auth(jwtManager))

			r.Get("/rates", currencyHandler.GetAllRates)
			r.Get("/rates/{currency}", currencyHandler.GetRate)
			r.Get("/rates/{currency}/history", currencyHandler.GetRateHistory)
		})
	})
}
