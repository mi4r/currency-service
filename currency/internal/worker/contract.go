package worker

import (
	"context"
	"time"

	"github.com/mi4r/currency-service/currency/internal/domain"
)

// RateRepository defines the interface for currency rate persistence operations.
// This interface is defined in the worker layer (consumer) following Dependency Inversion Principle.
type RateRepository interface {
	// Save saves or updates a currency rate.
	Save(ctx context.Context, rate *domain.CurrencyRate) error

	// SaveBatch saves multiple currency rates in a single transaction.
	SaveBatch(ctx context.Context, rates []*domain.CurrencyRate) error

	// GetExistingDates returns dates that already have rate data in the given range.
	GetExistingDates(ctx context.Context, from, to time.Time) ([]time.Time, error)
}
