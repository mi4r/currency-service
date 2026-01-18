package worker

import (
	"context"

	"github.com/mi4r/currency-service/currency/internal/domain"
)

// RateRepository defines the interface for currency rate persistence operations.
// This interface is defined in the worker layer (consumer) following Dependency Inversion Principle.
type RateRepository interface {
	// Save saves or updates a currency rate.
	Save(ctx context.Context, rate *domain.CurrencyRate) error

	// SaveBatch saves multiple currency rates in a single transaction.
	SaveBatch(ctx context.Context, rates []*domain.CurrencyRate) error
}
