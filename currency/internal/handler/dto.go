package handler

import "time"

// RateResponse represents a single rate response.
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

// DateFormat is the standard date format used in API.
const DateFormat = "2006-01-02"

// ParseDate parses a date string in the standard format.
func ParseDate(s string) (time.Time, error) {
	return time.Parse(DateFormat, s)
}

// FormatDate formats a time.Time to the standard date format.
func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}
