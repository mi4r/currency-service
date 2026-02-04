package repository

import (
	"context"
	"time"

	"github.com/mi4r/currency-service/currency/internal/domain"
)

// RateRepository defines the interface for currency rate data operations.
type RateRepository interface {
	// Save saves or updates a currency rate.
	Save(ctx context.Context, rate *domain.CurrencyRate) error

	// SaveBatch saves multiple currency rates in a single transaction.
	SaveBatch(ctx context.Context, rates []*domain.CurrencyRate) error

	// GetByDate retrieves a rate for a specific currency and date.
	GetByDate(ctx context.Context, targetCurrency string, date time.Time) (*domain.CurrencyRate, error)

	// GetByDateRange retrieves rates for a currency within a date range.
	GetByDateRange(ctx context.Context, targetCurrency string, from, to time.Time) ([]domain.CurrencyRate, error)

	// GetLatest retrieves the most recent rate for a currency.
	GetLatest(ctx context.Context, targetCurrency string) (*domain.CurrencyRate, error)

	// GetAllByDate retrieves all rates for a specific date.
	GetAllByDate(ctx context.Context, date time.Time) ([]domain.CurrencyRate, error)
}
