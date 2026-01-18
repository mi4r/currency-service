package domain

import (
	"errors"
	"time"
)

// Domain errors
var (
	ErrRateNotFound     = errors.New("rate not found")
	ErrInvalidDate      = errors.New("invalid date")
	ErrInvalidCurrency  = errors.New("invalid currency")
	ErrInvalidDateRange = errors.New("invalid date range")
)

// CurrencyRate represents a currency exchange rate.
type CurrencyRate struct {
	ID             int64     `json:"id"`
	RateDate       time.Time `json:"rate_date"`
	BaseCurrency   string    `json:"base_currency"`
	TargetCurrency string    `json:"target_currency"`
	Rate           float64   `json:"rate"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// RateHistory represents a single rate entry in history.
type RateHistory struct {
	Date time.Time `json:"date"`
	Rate float64   `json:"rate"`
}
