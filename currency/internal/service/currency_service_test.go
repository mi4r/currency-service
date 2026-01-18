package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mi4r/currency-service/currency/internal/domain"
)

type MockRateRepository struct {
	mock.Mock
}

func (m *MockRateRepository) GetByDate(ctx context.Context, targetCurrency string, date time.Time) (*domain.CurrencyRate, error) {
	args := m.Called(ctx, targetCurrency, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CurrencyRate), args.Error(1)
}

func (m *MockRateRepository) GetByDateRange(ctx context.Context, targetCurrency string, from, to time.Time) ([]domain.CurrencyRate, error) {
	args := m.Called(ctx, targetCurrency, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.CurrencyRate), args.Error(1)
}

func (m *MockRateRepository) GetLatest(ctx context.Context, targetCurrency string) (*domain.CurrencyRate, error) {
	args := m.Called(ctx, targetCurrency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CurrencyRate), args.Error(1)
}

func (m *MockRateRepository) GetAllByDate(ctx context.Context, date time.Time) ([]domain.CurrencyRate, error) {
	args := m.Called(ctx, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.CurrencyRate), args.Error(1)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestCurrencyService_GetRate(t *testing.T) {
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		currency     string
		date         time.Time
		setupMock    func(*MockRateRepository)
		expectedRate *domain.CurrencyRate
		expectedErr  error
	}{
		{
			name:     "successful rate retrieval",
			currency: "USD",
			date:     date,
			setupMock: func(m *MockRateRepository) {
				m.On("GetByDate", mock.Anything, "usd", date).Return(&domain.CurrencyRate{
					ID:             1,
					RateDate:       date,
					BaseCurrency:   "RUB",
					TargetCurrency: "usd",
					Rate:           0.012763493,
				}, nil)
			},
			expectedRate: &domain.CurrencyRate{
				ID:             1,
				RateDate:       date,
				BaseCurrency:   "RUB",
				TargetCurrency: "usd",
				Rate:           0.012763493,
			},
			expectedErr: nil,
		},
		{
			name:         "empty currency",
			currency:     "",
			date:         date,
			setupMock:    func(m *MockRateRepository) {},
			expectedRate: nil,
			expectedErr:  domain.ErrInvalidCurrency,
		},
		{
			name:     "rate not found",
			currency: "xyz",
			date:     date,
			setupMock: func(m *MockRateRepository) {
				m.On("GetByDate", mock.Anything, "xyz", date).Return(nil, domain.ErrRateNotFound)
			},
			expectedRate: nil,
			expectedErr:  domain.ErrRateNotFound,
		},
		{
			name:     "lowercase conversion",
			currency: "EUR",
			date:     date,
			setupMock: func(m *MockRateRepository) {
				m.On("GetByDate", mock.Anything, "eur", date).Return(&domain.CurrencyRate{
					TargetCurrency: "eur",
					Rate:           0.0109,
				}, nil)
			},
			expectedRate: &domain.CurrencyRate{
				TargetCurrency: "eur",
				Rate:           0.0109,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRateRepository)
			tt.setupMock(mockRepo)
			svc := NewCurrencyService(mockRepo, newTestLogger())

			rate, err := svc.GetRate(context.Background(), tt.currency, tt.date)

			if tt.expectedErr != nil {
				assert.Nil(t, rate)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRate, rate)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCurrencyService_GetRateHistory(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		currency        string
		from            time.Time
		to              time.Time
		setupMock       func(*MockRateRepository)
		expectedHistory []domain.RateHistory
		expectedErr     error
	}{
		{
			name:     "successful history retrieval",
			currency: "USD",
			from:     from,
			to:       to,
			setupMock: func(m *MockRateRepository) {
				m.On("GetByDateRange", mock.Anything, "usd", from, to).Return([]domain.CurrencyRate{
					{RateDate: from, Rate: 0.0125},
					{RateDate: from.AddDate(0, 0, 1), Rate: 0.0126},
					{RateDate: to, Rate: 0.0127},
				}, nil)
			},
			expectedHistory: []domain.RateHistory{
				{Date: from, Rate: 0.0125},
				{Date: from.AddDate(0, 0, 1), Rate: 0.0126},
				{Date: to, Rate: 0.0127},
			},
			expectedErr: nil,
		},
		{
			name:            "empty currency",
			currency:        "",
			from:            from,
			to:              to,
			setupMock:       func(m *MockRateRepository) {},
			expectedHistory: nil,
			expectedErr:     domain.ErrInvalidCurrency,
		},
		{
			name:            "invalid date range (from > to)",
			currency:        "USD",
			from:            to,
			to:              from,
			setupMock:       func(m *MockRateRepository) {},
			expectedHistory: nil,
			expectedErr:     domain.ErrInvalidDateRange,
		},
		{
			name:     "empty result",
			currency: "USD",
			from:     from,
			to:       to,
			setupMock: func(m *MockRateRepository) {
				m.On("GetByDateRange", mock.Anything, "usd", from, to).Return([]domain.CurrencyRate{}, nil)
			},
			expectedHistory: []domain.RateHistory{},
			expectedErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRateRepository)
			tt.setupMock(mockRepo)
			svc := NewCurrencyService(mockRepo, newTestLogger())

			history, err := svc.GetRateHistory(context.Background(), tt.currency, tt.from, tt.to)

			if tt.expectedErr != nil {
				assert.Nil(t, history)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedHistory, history)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCurrencyService_GetLatestRate(t *testing.T) {
	tests := []struct {
		name         string
		currency     string
		setupMock    func(*MockRateRepository)
		expectedRate *domain.CurrencyRate
		expectedErr  error
	}{
		{
			name:     "successful latest rate retrieval",
			currency: "USD",
			setupMock: func(m *MockRateRepository) {
				m.On("GetLatest", mock.Anything, "usd").Return(&domain.CurrencyRate{
					TargetCurrency: "usd",
					Rate:           0.012763493,
				}, nil)
			},
			expectedRate: &domain.CurrencyRate{
				TargetCurrency: "usd",
				Rate:           0.012763493,
			},
			expectedErr: nil,
		},
		{
			name:         "empty currency",
			currency:     "",
			setupMock:    func(m *MockRateRepository) {},
			expectedRate: nil,
			expectedErr:  domain.ErrInvalidCurrency,
		},
		{
			name:     "rate not found",
			currency: "xyz",
			setupMock: func(m *MockRateRepository) {
				m.On("GetLatest", mock.Anything, "xyz").Return(nil, domain.ErrRateNotFound)
			},
			expectedRate: nil,
			expectedErr:  domain.ErrRateNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRateRepository)
			tt.setupMock(mockRepo)
			svc := NewCurrencyService(mockRepo, newTestLogger())

			rate, err := svc.GetLatestRate(context.Background(), tt.currency)

			if tt.expectedErr != nil {
				assert.Nil(t, rate)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRate, rate)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCurrencyService_GetAllRates(t *testing.T) {
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		date          time.Time
		setupMock     func(*MockRateRepository)
		expectedRates map[string]float64
		expectedErr   error
	}{
		{
			name: "successful all rates retrieval",
			date: date,
			setupMock: func(m *MockRateRepository) {
				m.On("GetAllByDate", mock.Anything, date).Return([]domain.CurrencyRate{
					{TargetCurrency: "usd", Rate: 0.0127},
					{TargetCurrency: "eur", Rate: 0.0109},
				}, nil)
			},
			expectedRates: map[string]float64{"usd": 0.0127, "eur": 0.0109},
			expectedErr:   nil,
		},
		{
			name: "empty result",
			date: date,
			setupMock: func(m *MockRateRepository) {
				m.On("GetAllByDate", mock.Anything, date).Return([]domain.CurrencyRate{}, nil)
			},
			expectedRates: map[string]float64{},
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRateRepository)
			tt.setupMock(mockRepo)
			svc := NewCurrencyService(mockRepo, newTestLogger())

			rates, err := svc.GetAllRates(context.Background(), tt.date)

			if tt.expectedErr != nil {
				assert.Nil(t, rates)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRates, rates)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
