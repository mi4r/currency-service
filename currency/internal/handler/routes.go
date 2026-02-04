package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mi4r/currency-service/currency/internal/metrics"
)

// RegisterRoutes registers all routes for the currency service.
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(metricsMiddleware)

	r.Handle("/metrics", promhttp.Handler())

	r.Route("/internal/v1", func(r chi.Router) {
		r.Get("/health", h.Health)
		r.Get("/rates", h.GetAllRates)
		r.Get("/rates/{currency}", h.GetRate)
		r.Get("/rates/{currency}/history", h.GetRateHistory)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rec, r)

		duration := time.Since(start).Seconds()
		path := r.URL.Path
		method := r.Method
		status := fmt.Sprintf("%d", rec.statusCode)

		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	})
}
