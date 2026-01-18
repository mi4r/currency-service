package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/mi4r/currency-service/pkg/httputil"
)

// CurrencyHandler handles currency-related requests.
type CurrencyHandler struct {
	currencyClient CurrencyClient
	logger         *slog.Logger
}

// NewCurrencyHandler creates a new currency handler.
func NewCurrencyHandler(currencyClient CurrencyClient, logger *slog.Logger) *CurrencyHandler {
	return &CurrencyHandler{
		currencyClient: currencyClient,
		logger:         logger,
	}
}

// GetRate handles GET /api/v1/rates/{currency}
func (h *CurrencyHandler) GetRate(w http.ResponseWriter, r *http.Request) {
	currency := chi.URLParam(r, "currency")
	if currency == "" {
		httputil.BadRequest(w, "currency is required")
		return
	}

	date := r.URL.Query().Get("date")

	rate, err := h.currencyClient.GetRate(r.Context(), currency, date)
	if err != nil {
		h.handleError(w, err)
		return
	}

	httputil.Success(w, rate)
}

// GetRateHistory handles GET /api/v1/rates/{currency}/history
func (h *CurrencyHandler) GetRateHistory(w http.ResponseWriter, r *http.Request) {
	currency := chi.URLParam(r, "currency")
	if currency == "" {
		httputil.BadRequest(w, "currency is required")
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" || to == "" {
		httputil.BadRequest(w, "from and to dates are required")
		return
	}

	history, err := h.currencyClient.GetRateHistory(r.Context(), currency, from, to)
	if err != nil {
		h.handleError(w, err)
		return
	}

	httputil.Success(w, history)
}

// GetAllRates handles GET /api/v1/rates
func (h *CurrencyHandler) GetAllRates(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")

	rates, err := h.currencyClient.GetAllRates(r.Context(), date)
	if err != nil {
		h.handleError(w, err)
		return
	}

	httputil.Success(w, rates)
}

func (h *CurrencyHandler) handleError(w http.ResponseWriter, err error) {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "NOT_FOUND"):
		httputil.NotFound(w, "rate not found")
	case strings.Contains(errStr, "BAD_REQUEST"):
		httputil.BadRequest(w, errStr)
	default:
		h.logger.Error("currency client error", slog.String("error", errStr))
		httputil.InternalError(w, "internal server error")
	}
}
