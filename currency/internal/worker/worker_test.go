package worker

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mi4r/currency-service/currency/internal/domain"
)

type MockRateRepository struct {
	mock.Mock
}

func (m *MockRateRepository) Save(ctx context.Context, rate *domain.CurrencyRate) error {
	args := m.Called(ctx, rate)
	return args.Error(0)
}

func (m *MockRateRepository) SaveBatch(ctx context.Context, rates []*domain.CurrencyRate) error {
	args := m.Called(ctx, rates)
	return args.Error(0)
}

func (m *MockRateRepository) GetExistingDates(ctx context.Context, from, to time.Time) ([]time.Time, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).([]time.Time), args.Error(1)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestWorker_FetchNow(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*MockRateRepository)
		expectError bool
	}{
		{
			name: "successful fetch and save",
			setupMock: func(m *MockRateRepository) {
				m.On("SaveBatch", mock.Anything, mock.AnythingOfType("[]*domain.CurrencyRate")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "save batch error",
			setupMock: func(m *MockRateRepository) {
				m.On("SaveBatch", mock.Anything, mock.AnythingOfType("[]*domain.CurrencyRate")).Return(assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRateRepository)
			tt.setupMock(mockRepo)

			fetcher := NewFetcher("https://latest.currency-api.pages.dev/v1/currencies/rub.json")
			w := NewWorker(fetcher, mockRepo, newTestLogger(), 24*time.Hour, 1, 1*time.Second, "", 0)

			err := w.FetchNow(context.Background())

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWorker_StartAndStop(t *testing.T) {
	mockRepo := new(MockRateRepository)
	mockRepo.On("SaveBatch", mock.Anything, mock.AnythingOfType("[]*domain.CurrencyRate")).Return(nil).Maybe()

	fetcher := NewFetcher("https://latest.currency-api.pages.dev/v1/currencies/rub.json")
	w := NewWorker(fetcher, mockRepo, newTestLogger(), 1*time.Hour, 1, 1*time.Second, "", 0)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		w.Start(ctx)
		close(done)
	}()

	// Give the worker time to start and perform initial fetch
	time.Sleep(100 * time.Millisecond)

	// Stop the worker
	cancel()

	select {
	case <-done:
		// Worker stopped successfully
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not stop in time")
	}
}
