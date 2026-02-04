package handler

import (
	"context"
	"time"

	"github.com/mi4r/currency-service/currency/internal/domain"
)

// CurrencyService defines the interface for currency rate operations.
// This interface is defined in the handler layer (consumer) following Dependency Inversion Principle.
type CurrencyService interface {
	// GetRate retrieves a rate for a specific currency and date.
	GetRate(ctx context.Context, currency string, date time.Time) (*domain.CurrencyRate, error)

	// GetRateHistory retrieves rates for a currency within a date range.
	GetRateHistory(ctx context.Context, currency string, from, to time.Time) ([]domain.RateHistory, error)

	// GetLatestRate retrieves the most recent rate for a currency.
	GetLatestRate(ctx context.Context, currency string) (*domain.CurrencyRate, error)

	// GetAllRates retrieves all rates for a specific date.
	GetAllRates(ctx context.Context, date time.Time) (map[string]float64, error)
}
