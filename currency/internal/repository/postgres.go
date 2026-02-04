package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mi4r/currency-service/currency/internal/domain"
)

// PostgresRepository implements service.RateRepository and worker.RateRepository interfaces.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// Save saves or updates a currency rate.
func (r *PostgresRepository) Save(ctx context.Context, rate *domain.CurrencyRate) error {
	query := `
		INSERT INTO currency_rates (rate_date, base_currency, target_currency, rate)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (rate_date, base_currency, target_currency)
		DO UPDATE SET rate = EXCLUDED.rate, updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		rate.RateDate,
		rate.BaseCurrency,
		rate.TargetCurrency,
		rate.Rate,
	).Scan(&rate.ID, &rate.CreatedAt, &rate.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to save rate: %w", err)
	}

	return nil
}

// SaveBatch saves multiple currency rates in a single transaction.
func (r *PostgresRepository) SaveBatch(ctx context.Context, rates []*domain.CurrencyRate) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO currency_rates (rate_date, base_currency, target_currency, rate)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (rate_date, base_currency, target_currency)
		DO UPDATE SET rate = EXCLUDED.rate, updated_at = NOW()
	`

	for _, rate := range rates {
		_, err := tx.Exec(ctx, query,
			rate.RateDate,
			rate.BaseCurrency,
			rate.TargetCurrency,
			rate.Rate,
		)
		if err != nil {
			return fmt.Errorf("failed to save rate for %s: %w", rate.TargetCurrency, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetExistingDates returns distinct dates that have rate data in the given range.
func (r *PostgresRepository) GetExistingDates(ctx context.Context, from, to time.Time) ([]time.Time, error) {
	query := `
		SELECT DISTINCT rate_date
		FROM currency_rates
		WHERE rate_date >= $1 AND rate_date <= $2
		ORDER BY rate_date ASC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query existing dates: %w", err)
	}
	defer rows.Close()

	var dates []time.Time
	for rows.Next() {
		var date time.Time
		if err := rows.Scan(&date); err != nil {
			return nil, fmt.Errorf("failed to scan date: %w", err)
		}
		dates = append(dates, date)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating dates: %w", err)
	}

	return dates, nil
}

// GetByDate retrieves a rate for a specific currency and date.
func (r *PostgresRepository) GetByDate(ctx context.Context, targetCurrency string, date time.Time) (*domain.CurrencyRate, error) {
	query := `
		SELECT id, rate_date, base_currency, target_currency, rate, created_at, updated_at
		FROM currency_rates
		WHERE target_currency = $1 AND rate_date = $2
	`

	rate := &domain.CurrencyRate{}
	err := r.pool.QueryRow(ctx, query, targetCurrency, date).Scan(
		&rate.ID,
		&rate.RateDate,
		&rate.BaseCurrency,
		&rate.TargetCurrency,
		&rate.Rate,
		&rate.CreatedAt,
		&rate.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRateNotFound
		}
		return nil, fmt.Errorf("failed to get rate: %w", err)
	}

	return rate, nil
}

// GetByDateRange retrieves rates for a currency within a date range.
func (r *PostgresRepository) GetByDateRange(ctx context.Context, targetCurrency string, from, to time.Time) ([]domain.CurrencyRate, error) {
	query := `
		SELECT id, rate_date, base_currency, target_currency, rate, created_at, updated_at
		FROM currency_rates
		WHERE target_currency = $1 AND rate_date >= $2 AND rate_date <= $3
		ORDER BY rate_date ASC
	`

	rows, err := r.pool.Query(ctx, query, targetCurrency, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query rates: %w", err)
	}
	defer rows.Close()

	var rates []domain.CurrencyRate
	for rows.Next() {
		var rate domain.CurrencyRate
		err := rows.Scan(
			&rate.ID,
			&rate.RateDate,
			&rate.BaseCurrency,
			&rate.TargetCurrency,
			&rate.Rate,
			&rate.CreatedAt,
			&rate.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rate: %w", err)
		}
		rates = append(rates, rate)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return rates, nil
}

// GetLatest retrieves the most recent rate for a currency.
func (r *PostgresRepository) GetLatest(ctx context.Context, targetCurrency string) (*domain.CurrencyRate, error) {
	query := `
		SELECT id, rate_date, base_currency, target_currency, rate, created_at, updated_at
		FROM currency_rates
		WHERE target_currency = $1
		ORDER BY rate_date DESC
		LIMIT 1
	`

	rate := &domain.CurrencyRate{}
	err := r.pool.QueryRow(ctx, query, targetCurrency).Scan(
		&rate.ID,
		&rate.RateDate,
		&rate.BaseCurrency,
		&rate.TargetCurrency,
		&rate.Rate,
		&rate.CreatedAt,
		&rate.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRateNotFound
		}
		return nil, fmt.Errorf("failed to get latest rate: %w", err)
	}

	return rate, nil
}

// GetAllByDate retrieves all rates for a specific date.
func (r *PostgresRepository) GetAllByDate(ctx context.Context, date time.Time) ([]domain.CurrencyRate, error) {
	query := `
		SELECT id, rate_date, base_currency, target_currency, rate, created_at, updated_at
		FROM currency_rates
		WHERE rate_date = $1
		ORDER BY target_currency ASC
	`

	rows, err := r.pool.Query(ctx, query, date)
	if err != nil {
		return nil, fmt.Errorf("failed to query rates: %w", err)
	}
	defer rows.Close()

	var rates []domain.CurrencyRate
	for rows.Next() {
		var rate domain.CurrencyRate
		err := rows.Scan(
			&rate.ID,
			&rate.RateDate,
			&rate.BaseCurrency,
			&rate.TargetCurrency,
			&rate.Rate,
			&rate.CreatedAt,
			&rate.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rate: %w", err)
		}
		rates = append(rates, rate)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return rates, nil
}
