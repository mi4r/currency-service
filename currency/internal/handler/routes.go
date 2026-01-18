package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// RegisterRoutes registers all routes for the currency service.
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Route("/internal/v1", func(r chi.Router) {
		r.Get("/health", h.Health)
		r.Get("/rates", h.GetAllRates)
		r.Get("/rates/{currency}", h.GetRate)
		r.Get("/rates/{currency}/history", h.GetRateHistory)
	})
}
