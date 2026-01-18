package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/mi4r/currency-service/currency/internal/domain"
	"github.com/mi4r/currency-service/pkg/httputil"
)

// Handler handles HTTP requests for currency rates.
type Handler struct {
	service CurrencyService
	logger  *slog.Logger
}

// NewHandler creates a new handler instance.
func NewHandler(service CurrencyService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetRate handles GET /internal/v1/rates/{currency}
func (h *Handler) GetRate(w http.ResponseWriter, r *http.Request) {
	currency := chi.URLParam(r, "currency")
	if currency == "" {
		httputil.BadRequest(w, "currency is required")
		return
	}

	dateStr := r.URL.Query().Get("date")
	var date time.Time
	var err error

	if dateStr != "" {
		date, err = ParseDate(dateStr)
		if err != nil {
			httputil.BadRequest(w, "invalid date format, use YYYY-MM-DD")
			return
		}
	} else {
		// Get latest rate if no date specified
		rate, err := h.service.GetLatestRate(r.Context(), currency)
		if err != nil {
			h.handleError(w, err)
			return
		}

		httputil.Success(w, RateResponse{
			Date:           FormatDate(rate.RateDate),
			BaseCurrency:   rate.BaseCurrency,
			TargetCurrency: rate.TargetCurrency,
			Rate:           rate.Rate,
		})
		return
	}

	rate, err := h.service.GetRate(r.Context(), currency, date)
	if err != nil {
		h.handleError(w, err)
		return
	}

	httputil.Success(w, RateResponse{
		Date:           FormatDate(rate.RateDate),
		BaseCurrency:   rate.BaseCurrency,
		TargetCurrency: rate.TargetCurrency,
		Rate:           rate.Rate,
	})
}

// GetRateHistory handles GET /internal/v1/rates/{currency}/history
func (h *Handler) GetRateHistory(w http.ResponseWriter, r *http.Request) {
	currency := chi.URLParam(r, "currency")
	if currency == "" {
		httputil.BadRequest(w, "currency is required")
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		httputil.BadRequest(w, "from and to dates are required")
		return
	}

	from, err := ParseDate(fromStr)
	if err != nil {
		httputil.BadRequest(w, "invalid from date format, use YYYY-MM-DD")
		return
	}

	to, err := ParseDate(toStr)
	if err != nil {
		httputil.BadRequest(w, "invalid to date format, use YYYY-MM-DD")
		return
	}

	history, err := h.service.GetRateHistory(r.Context(), currency, from, to)
	if err != nil {
		h.handleError(w, err)
		return
	}

	items := make([]RateHistoryItem, len(history))
	for i, h := range history {
		items[i] = RateHistoryItem{
			Date: FormatDate(h.Date),
			Rate: h.Rate,
		}
	}

	httputil.Success(w, RateHistoryResponse{
		BaseCurrency:   "RUB",
		TargetCurrency: currency,
		From:           fromStr,
		To:             toStr,
		Rates:          items,
	})
}

// GetAllRates handles GET /internal/v1/rates
func (h *Handler) GetAllRates(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	var date time.Time
	var err error

	if dateStr != "" {
		date, err = ParseDate(dateStr)
		if err != nil {
			httputil.BadRequest(w, "invalid date format, use YYYY-MM-DD")
			return
		}
	} else {
		date = time.Now()
	}

	rates, err := h.service.GetAllRates(r.Context(), date)
	if err != nil {
		h.handleError(w, err)
		return
	}

	httputil.Success(w, AllRatesResponse{
		Date:         FormatDate(date),
		BaseCurrency: "RUB",
		Rates:        rates,
	})
}

// Health handles GET /internal/v1/health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	httputil.Success(w, map[string]string{"status": "ok"})
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrRateNotFound):
		httputil.NotFound(w, "rate not found")
	case errors.Is(err, domain.ErrInvalidCurrency):
		httputil.BadRequest(w, "invalid currency")
	case errors.Is(err, domain.ErrInvalidDateRange):
		httputil.BadRequest(w, "invalid date range")
	default:
		h.logger.Error("internal error", slog.String("error", err.Error()))
		httputil.InternalError(w, "internal server error")
	}
}
