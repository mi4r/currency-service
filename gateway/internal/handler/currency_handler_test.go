package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mi4r/currency-service/gateway/internal/client"
	"github.com/mi4r/currency-service/pkg/logger"
)

type MockCurrencyClient struct {
	mock.Mock
}

func (m *MockCurrencyClient) GetRate(ctx context.Context, currency string, date string) (*client.RateResponse, error) {
	args := m.Called(ctx, currency, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.RateResponse), args.Error(1)
}

func (m *MockCurrencyClient) GetRateHistory(ctx context.Context, currency, from, to string) (*client.RateHistoryResponse, error) {
	args := m.Called(ctx, currency, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.RateHistoryResponse), args.Error(1)
}

func (m *MockCurrencyClient) GetAllRates(ctx context.Context, date string) (*client.AllRatesResponse, error) {
	args := m.Called(ctx, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.AllRatesResponse), args.Error(1)
}

func TestCurrencyHandler_GetRate(t *testing.T) {
	tests := []struct {
		name           string
		currency       string
		dateParam      string
		setupMock      func(*MockCurrencyClient)
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:      "successful rate retrieval",
			currency:  "usd",
			dateParam: "2026-01-15",
			setupMock: func(m *MockCurrencyClient) {
				m.On("GetRate", mock.Anything, "usd", "2026-01-15").Return(&client.RateResponse{
					Date:           "2026-01-15",
					BaseCurrency:   "RUB",
					TargetCurrency: "usd",
					Rate:           0.012763493,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "2026-01-15", body["date"])
				assert.Equal(t, "RUB", body["base_currency"])
				assert.Equal(t, "usd", body["target_currency"])
				assert.Equal(t, 0.012763493, body["rate"])
			},
		},
		{
			name:      "rate without date parameter",
			currency:  "usd",
			dateParam: "",
			setupMock: func(m *MockCurrencyClient) {
				m.On("GetRate", mock.Anything, "usd", "").Return(&client.RateResponse{
					Date:           "2026-01-17",
					BaseCurrency:   "RUB",
					TargetCurrency: "usd",
					Rate:           0.012763493,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "2026-01-17", body["date"])
			},
		},
		{
			name:      "rate not found",
			currency:  "xyz",
			dateParam: "2026-01-15",
			setupMock: func(m *MockCurrencyClient) {
				m.On("GetRate", mock.Anything, "xyz", "2026-01-15").Return(nil, errors.New("NOT_FOUND: rate not found"))
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "rate not found", body["error"])
			},
		},
		{
			name:      "bad request from currency service",
			currency:  "usd",
			dateParam: "invalid",
			setupMock: func(m *MockCurrencyClient) {
				m.On("GetRate", mock.Anything, "usd", "invalid").Return(nil, errors.New("BAD_REQUEST: invalid date format"))
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "BAD_REQUEST")
			},
		},
		{
			name:      "internal error",
			currency:  "usd",
			dateParam: "2026-01-15",
			setupMock: func(m *MockCurrencyClient) {
				m.On("GetRate", mock.Anything, "usd", "2026-01-15").Return(nil, errors.New("connection refused"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "internal server error", body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockCurrencyClient)
			tt.setupMock(mockClient)
			h := NewCurrencyHandler(mockClient, logger.New("error"))

			r := chi.NewRouter()
			r.Get("/api/v1/rates/{currency}", h.GetRate)

			url := "/api/v1/rates/" + tt.currency
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
			mockClient.AssertExpectations(t)
		})
	}
}

func TestCurrencyHandler_GetRateHistory(t *testing.T) {
	tests := []struct {
		name           string
		currency       string
		fromParam      string
		toParam        string
		setupMock      func(*MockCurrencyClient)
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:      "successful history retrieval",
			currency:  "usd",
			fromParam: "2026-01-01",
			toParam:   "2026-01-03",
			setupMock: func(m *MockCurrencyClient) {
				m.On("GetRateHistory", mock.Anything, "usd", "2026-01-01", "2026-01-03").Return(&client.RateHistoryResponse{
					BaseCurrency:   "RUB",
					TargetCurrency: "usd",
					From:           "2026-01-01",
					To:             "2026-01-03",
					Rates: []client.RateHistoryItem{
						{Date: "2026-01-01", Rate: 0.0125},
						{Date: "2026-01-03", Rate: 0.0127},
					},
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
			setupMock:      func(m *MockCurrencyClient) {},
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
			setupMock:      func(m *MockCurrencyClient) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "from and to dates are required", body["error"])
			},
		},
		{
			name:      "currency service error",
			currency:  "usd",
			fromParam: "2026-01-01",
			toParam:   "2026-01-03",
			setupMock: func(m *MockCurrencyClient) {
				m.On("GetRateHistory", mock.Anything, "usd", "2026-01-01", "2026-01-03").Return(nil, errors.New("connection refused"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "internal server error", body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockCurrencyClient)
			tt.setupMock(mockClient)
			h := NewCurrencyHandler(mockClient, logger.New("error"))

			r := chi.NewRouter()
			r.Get("/api/v1/rates/{currency}/history", h.GetRateHistory)

			url := "/api/v1/rates/" + tt.currency + "/history?"
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
			mockClient.AssertExpectations(t)
		})
	}
}

func TestCurrencyHandler_GetAllRates(t *testing.T) {
	tests := []struct {
		name           string
		dateParam      string
		setupMock      func(*MockCurrencyClient)
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:      "successful all rates retrieval",
			dateParam: "2026-01-15",
			setupMock: func(m *MockCurrencyClient) {
				m.On("GetAllRates", mock.Anything, "2026-01-15").Return(&client.AllRatesResponse{
					Date:         "2026-01-15",
					BaseCurrency: "RUB",
					Rates: map[string]float64{
						"usd": 0.0127,
						"eur": 0.0109,
					},
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
			name:      "all rates without date",
			dateParam: "",
			setupMock: func(m *MockCurrencyClient) {
				m.On("GetAllRates", mock.Anything, "").Return(&client.AllRatesResponse{
					Date:         "2026-01-17",
					BaseCurrency: "RUB",
					Rates:        map[string]float64{"usd": 0.0127},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "2026-01-17", body["date"])
			},
		},
		{
			name:      "currency service error",
			dateParam: "2026-01-15",
			setupMock: func(m *MockCurrencyClient) {
				m.On("GetAllRates", mock.Anything, "2026-01-15").Return(nil, errors.New("connection refused"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "internal server error", body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockCurrencyClient)
			tt.setupMock(mockClient)
			h := NewCurrencyHandler(mockClient, logger.New("error"))

			r := chi.NewRouter()
			r.Get("/api/v1/rates", h.GetAllRates)

			url := "/api/v1/rates"
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
			mockClient.AssertExpectations(t)
		})
	}
}
