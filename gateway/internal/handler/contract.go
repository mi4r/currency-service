package handler

import (
	"context"

	"github.com/mi4r/currency-service/gateway/internal/client"
	"github.com/mi4r/currency-service/gateway/internal/service"
)

// AuthService defines the interface for authentication operations.
// This interface is defined in the handler layer (consumer) following Dependency Inversion Principle.
type AuthService interface {
	// Login authenticates a user and returns a JWT token.
	Login(ctx context.Context, username, password string) (*service.AuthResult, error)
}

// CurrencyClient defines the interface for currency service operations.
// This interface is defined in the handler layer (consumer) following Dependency Inversion Principle.
type CurrencyClient interface {
	// GetRate retrieves a rate for a specific currency and optional date.
	GetRate(ctx context.Context, currency string, date string) (*client.RateResponse, error)

	// GetRateHistory retrieves rate history for a currency within a date range.
	GetRateHistory(ctx context.Context, currency, from, to string) (*client.RateHistoryResponse, error)

	// GetAllRates retrieves all rates for a specific date.
	GetAllRates(ctx context.Context, date string) (*client.AllRatesResponse, error)
}
