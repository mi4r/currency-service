package service

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/mi4r/currency-service/currency/internal/domain"
)

// CurrencyServiceImpl implements handler.CurrencyService interface.
type CurrencyServiceImpl struct {
	repo   RateRepository
	logger *slog.Logger
}

// NewCurrencyService creates a new currency service.
func NewCurrencyService(repo RateRepository, logger *slog.Logger) *CurrencyServiceImpl {
	return &CurrencyServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

// GetRate retrieves a rate for a specific currency and date.
func (s *CurrencyServiceImpl) GetRate(ctx context.Context, currency string, date time.Time) (*domain.CurrencyRate, error) {
	currency = strings.ToLower(currency)

	if currency == "" {
		return nil, domain.ErrInvalidCurrency
	}

	rate, err := s.repo.GetByDate(ctx, currency, date)
	if err != nil {
		s.logger.Error("failed to get rate",
			slog.String("currency", currency),
			slog.Time("date", date),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return rate, nil
}

// GetRateHistory retrieves rates for a currency within a date range.
func (s *CurrencyServiceImpl) GetRateHistory(ctx context.Context, currency string, from, to time.Time) ([]domain.RateHistory, error) {
	currency = strings.ToLower(currency)

	if currency == "" {
		return nil, domain.ErrInvalidCurrency
	}

	if from.After(to) {
		return nil, domain.ErrInvalidDateRange
	}

	rates, err := s.repo.GetByDateRange(ctx, currency, from, to)
	if err != nil {
		s.logger.Error("failed to get rate history",
			slog.String("currency", currency),
			slog.Time("from", from),
			slog.Time("to", to),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	history := make([]domain.RateHistory, len(rates))
	for i, rate := range rates {
		history[i] = domain.RateHistory{
			Date: rate.RateDate,
			Rate: rate.Rate,
		}
	}

	return history, nil
}

// GetLatestRate retrieves the most recent rate for a currency.
func (s *CurrencyServiceImpl) GetLatestRate(ctx context.Context, currency string) (*domain.CurrencyRate, error) {
	currency = strings.ToLower(currency)

	if currency == "" {
		return nil, domain.ErrInvalidCurrency
	}

	rate, err := s.repo.GetLatest(ctx, currency)
	if err != nil {
		s.logger.Error("failed to get latest rate",
			slog.String("currency", currency),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return rate, nil
}

// GetAllRates retrieves all rates for a specific date.
func (s *CurrencyServiceImpl) GetAllRates(ctx context.Context, date time.Time) (map[string]float64, error) {
	rates, err := s.repo.GetAllByDate(ctx, date)
	if err != nil {
		s.logger.Error("failed to get all rates",
			slog.Time("date", date),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	result := make(map[string]float64)
	for _, rate := range rates {
		result[rate.TargetCurrency] = rate.Rate
	}

	return result, nil
}
