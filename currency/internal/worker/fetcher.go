package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mi4r/currency-service/currency/internal/domain"
)

// APIResponse represents the currency API response structure.
type APIResponse struct {
	Date string             `json:"date"`
	RUB  map[string]float64 `json:"rub"`
}

// Fetcher fetches currency rates from external API.
type Fetcher struct {
	apiURL     string
	httpClient *http.Client
}

// NewFetcher creates a new currency rate fetcher.
func NewFetcher(apiURL string) *Fetcher {
	return &Fetcher{
		apiURL: apiURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Fetch retrieves currency rates from the API.
func (f *Fetcher) Fetch(ctx context.Context) (*APIResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &apiResp, nil
}

// FetchRate retrieves a specific currency rate.
func (f *Fetcher) FetchRate(ctx context.Context, targetCurrency string) (*domain.CurrencyRate, error) {
	apiResp, err := f.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	rate, ok := apiResp.RUB[targetCurrency]
	if !ok {
		return nil, fmt.Errorf("currency %s not found in API response", targetCurrency)
	}

	rateDate, err := time.Parse("2006-01-02", apiResp.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}

	return &domain.CurrencyRate{
		RateDate:       rateDate,
		BaseCurrency:   "RUB",
		TargetCurrency: targetCurrency,
		Rate:           rate,
	}, nil
}

// FetchAllRates retrieves all currency rates.
func (f *Fetcher) FetchAllRates(ctx context.Context) ([]*domain.CurrencyRate, error) {
	apiResp, err := f.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	rateDate, err := time.Parse("2006-01-02", apiResp.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}

	rates := make([]*domain.CurrencyRate, 0, len(apiResp.RUB))
	for currency, rate := range apiResp.RUB {
		rates = append(rates, &domain.CurrencyRate{
			RateDate:       rateDate,
			BaseCurrency:   "RUB",
			TargetCurrency: currency,
			Rate:           rate,
		})
	}

	return rates, nil
}
