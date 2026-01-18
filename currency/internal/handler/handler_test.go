package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mi4r/currency-service/currency/internal/domain"
	"github.com/mi4r/currency-service/pkg/logger"
)

type MockCurrencyService struct {
	mock.Mock
}

func (m *MockCurrencyService) GetRate(ctx context.Context, currency string, date time.Time) (*domain.CurrencyRate, error) {
	args := m.Called(ctx, currency, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CurrencyRate), args.Error(1)
}

func (m *MockCurrencyService) GetRateHistory(ctx context.Context, currency string, from, to time.Time) ([]domain.RateHistory, error) {
	args := m.Called(ctx, currency, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.RateHistory), args.Error(1)
}

func (m *MockCurrencyService) GetLatestRate(ctx context.Context, currency string) (*domain.CurrencyRate, error) {
	args := m.Called(ctx, currency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CurrencyRate), args.Error(1)
}

func (m *MockCurrencyService) GetAllRates(ctx context.Context, date time.Time) (map[string]float64, error) {
	args := m.Called(ctx, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]float64), args.Error(1)
}

func TestHandler_GetRate(t *testing.T) {
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		currency       string
		dateParam      string
		setupMock      func(*MockCurrencyService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:      "successful rate retrieval with date",
			currency:  "usd",
			dateParam: "2026-01-15",
			setupMock: func(m *MockCurrencyService) {
				m.On("GetRate", mock.Anything, "usd", date).Return(&domain.CurrencyRate{
					RateDate:       date,
					BaseCurrency:   "RUB",
					TargetCurrency: "usd",
					Rate:           0.012763493,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"date":            "2026-01-15",
				"base_currency":   "RUB",
				"target_currency": "usd",
				"rate":            0.012763493,
			},
		},
		{
			name:      "successful latest rate retrieval without date",
			currency:  "usd",
			dateParam: "",
			setupMock: func(m *MockCurrencyService) {
				m.On("GetLatestRate", mock.Anything, "usd").Return(&domain.CurrencyRate{
					RateDate:       date,
					BaseCurrency:   "RUB",
					TargetCurrency: "usd",
					Rate:           0.012763493,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"date":            "2026-01-15",
				"base_currency":   "RUB",
				"target_currency": "usd",
				"rate":            0.012763493,
			},
		},
		{
			name:           "invalid date format",
			currency:       "usd",
			dateParam:      "invalid-date",
			setupMock:      func(m *MockCurrencyService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "invalid date format, use YYYY-MM-DD",
				"code":  "BAD_REQUEST",
			},
		},
		{
			name:      "rate not found",
			currency:  "xyz",
			dateParam: "2026-01-15",
			setupMock: func(m *MockCurrencyService) {
				m.On("GetRate", mock.Anything, "xyz", date).Return(nil, domain.ErrRateNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "rate not found",
				"code":  "NOT_FOUND",
			},
		},
		{
			name:      "invalid currency",
			currency:  "usd",
			dateParam: "2026-01-15",
			setupMock: func(m *MockCurrencyService) {
				m.On("GetRate", mock.Anything, "usd", date).Return(nil, domain.ErrInvalidCurrency)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "invalid currency",
				"code":  "BAD_REQUEST",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockCurrencyService)
			tt.setupMock(mockSvc)
			h := NewHandler(mockSvc, logger.New("error"))

			r := chi.NewRouter()
			r.Get("/internal/v1/rates/{currency}", h.GetRate)

			url := "/internal/v1/rates/" + tt.currency
			if tt.dateParam != "" {
				url += "?date=" + tt.dateParam
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var body map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &body)
			require.NoError(t, err)

			for key, expected := range tt.expectedBody {
				assert.Equal(t, expected, body[key], "mismatch for key: %s", key)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_GetRateHistory(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		currency       string
		fromParam      string
		toParam        string
		setupMock      func(*MockCurrencyService)
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:      "successful history retrieval",
			currency:  "usd",
			fromParam: "2026-01-01",
			toParam:   "2026-01-03",
			setupMock: func(m *MockCurrencyService) {
				m.On("GetRateHistory", mock.Anything, "usd", from, to).Return([]domain.RateHistory{
					{Date: from, Rate: 0.0125},
					{Date: to, Rate: 0.0127},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "RUB", body["base_currency"])
				assert.Equal(t, "usd", body["target_currency"])
				rates := body["rates"].([]interface{})
				assert.Len(t, rates, 2)
			},
		},
		{
			name:           "missing from parameter",
			currency:       "usd",
			fromParam:      "",
			toParam:        "2026-01-03",
			setupMock:      func(m *MockCurrencyService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "from and to dates are required", body["error"])
			},
		},
		{
			name:           "missing to parameter",
			currency:       "usd",
			fromParam:      "2026-01-01",
			toParam:        "",
			setupMock:      func(m *MockCurrencyService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "from and to dates are required", body["error"])
			},
		},
		{
			name:           "invalid from date format",
			currency:       "usd",
			fromParam:      "invalid",
			toParam:        "2026-01-03",
			setupMock:      func(m *MockCurrencyService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "invalid from date format, use YYYY-MM-DD", body["error"])
			},
		},
		{
			name:           "invalid to date format",
			currency:       "usd",
			fromParam:      "2026-01-01",
			toParam:        "invalid",
			setupMock:      func(m *MockCurrencyService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "invalid to date format, use YYYY-MM-DD", body["error"])
			},
		},
		{
			name:      "invalid date range",
			currency:  "usd",
			fromParam: "2026-01-01",
			toParam:   "2026-01-03",
			setupMock: func(m *MockCurrencyService) {
				m.On("GetRateHistory", mock.Anything, "usd", from, to).Return(nil, domain.ErrInvalidDateRange)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "invalid date range", body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockCurrencyService)
			tt.setupMock(mockSvc)
			h := NewHandler(mockSvc, logger.New("error"))

			r := chi.NewRouter()
			r.Get("/internal/v1/rates/{currency}/history", h.GetRateHistory)

			url := "/internal/v1/rates/" + tt.currency + "/history?"
			if tt.fromParam != "" {
				url += "from=" + tt.fromParam + "&"
			}
			if tt.toParam != "" {
				url += "to=" + tt.toParam
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var body map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &body)
			require.NoError(t, err)

			tt.checkResponse(t, body)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_GetAllRates(t *testing.T) {
	tests := []struct {
		name           string
		dateParam      string
		setupMock      func(*MockCurrencyService)
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:      "successful all rates retrieval with date",
			dateParam: "2026-01-15",
			setupMock: func(m *MockCurrencyService) {
				date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
				m.On("GetAllRates", mock.Anything, date).Return(map[string]float64{
					"usd": 0.0127,
					"eur": 0.0109,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "2026-01-15", body["date"])
				assert.Equal(t, "RUB", body["base_currency"])
				rates := body["rates"].(map[string]interface{})
				assert.Equal(t, 0.0127, rates["usd"])
				assert.Equal(t, 0.0109, rates["eur"])
			},
		},
		{
			name:           "invalid date format",
			dateParam:      "invalid-date",
			setupMock:      func(m *MockCurrencyService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "invalid date format, use YYYY-MM-DD", body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockCurrencyService)
			tt.setupMock(mockSvc)
			h := NewHandler(mockSvc, logger.New("error"))

			r := chi.NewRouter()
			r.Get("/internal/v1/rates", h.GetAllRates)

			url := "/internal/v1/rates"
			if tt.dateParam != "" {
				url += "?date=" + tt.dateParam
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var body map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &body)
			require.NoError(t, err)

			tt.checkResponse(t, body)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Health(t *testing.T) {
	mockSvc := new(MockCurrencyService)
	h := NewHandler(mockSvc, logger.New("error"))

	req := httptest.NewRequest(http.MethodGet, "/internal/v1/health", nil)
	rr := httptest.NewRecorder()

	h.Health(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var body map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "ok", body["status"])
}
