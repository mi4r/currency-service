package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// RateResponse represents a rate response from currency service.
type RateResponse struct {
	Date           string  `json:"date"`
	BaseCurrency   string  `json:"base_currency"`
	TargetCurrency string  `json:"target_currency"`
	Rate           float64 `json:"rate"`
}

// RateHistoryItem represents a single item in rate history.
type RateHistoryItem struct {
	Date string  `json:"date"`
	Rate float64 `json:"rate"`
}

// RateHistoryResponse represents the rate history response.
type RateHistoryResponse struct {
	BaseCurrency   string            `json:"base_currency"`
	TargetCurrency string            `json:"target_currency"`
	From           string            `json:"from"`
	To             string            `json:"to"`
	Rates          []RateHistoryItem `json:"rates"`
}

// AllRatesResponse represents all rates for a date.
type AllRatesResponse struct {
	Date         string             `json:"date"`
	BaseCurrency string             `json:"base_currency"`
	Rates        map[string]float64 `json:"rates"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// CurrencyClient is an HTTP client for the currency service.
type CurrencyClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewCurrencyClient creates a new currency service client.
func NewCurrencyClient(baseURL string, timeout time.Duration) *CurrencyClient {
	return &CurrencyClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// GetRate retrieves a rate for a specific currency and optional date.
func (c *CurrencyClient) GetRate(ctx context.Context, currency string, date string) (*RateResponse, error) {
	u, err := url.Parse(fmt.Sprintf("%s/internal/v1/rates/%s", c.baseURL, currency))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if date != "" {
		q := u.Query()
		q.Set("date", date)
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var result RateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetRateHistory retrieves rate history for a currency within a date range.
func (c *CurrencyClient) GetRateHistory(ctx context.Context, currency, from, to string) (*RateHistoryResponse, error) {
	u, err := url.Parse(fmt.Sprintf("%s/internal/v1/rates/%s/history", c.baseURL, currency))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	q.Set("from", from)
	q.Set("to", to)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var result RateHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetAllRates retrieves all rates for a specific date.
func (c *CurrencyClient) GetAllRates(ctx context.Context, date string) (*AllRatesResponse, error) {
	u, err := url.Parse(fmt.Sprintf("%s/internal/v1/rates", c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if date != "" {
		q := u.Query()
		q.Set("date", date)
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var result AllRatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// HealthCheck checks if the currency service is healthy.
func (c *CurrencyClient) HealthCheck(ctx context.Context) error {
	u := fmt.Sprintf("%s/internal/v1/health", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *CurrencyClient) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return fmt.Errorf("%s: %s", errResp.Code, errResp.Error)
}
