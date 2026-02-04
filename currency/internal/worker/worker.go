package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/mi4r/currency-service/currency/internal/metrics"
)

// Worker handles periodic currency rate fetching.
type Worker struct {
	fetcher          *Fetcher
	repo             RateRepository
	logger           *slog.Logger
	interval         time.Duration
	retryAttempts    int
	retryDelay       time.Duration
	historicalAPIURL string
	backfillDays     int
	stopCh           chan struct{}
}

// NewWorker creates a new worker instance.
func NewWorker(
	fetcher *Fetcher,
	repo RateRepository,
	logger *slog.Logger,
	interval time.Duration,
	retryAttempts int,
	retryDelay time.Duration,
	historicalAPIURL string,
	backfillDays int,
) *Worker {
	return &Worker{
		fetcher:          fetcher,
		repo:             repo,
		logger:           logger,
		interval:         interval,
		retryAttempts:    retryAttempts,
		retryDelay:       retryDelay,
		historicalAPIURL: historicalAPIURL,
		backfillDays:     backfillDays,
		stopCh:           make(chan struct{}),
	}
}

// Start begins the periodic rate fetching.
func (w *Worker) Start(ctx context.Context) {
	w.logger.Info("starting currency rate worker",
		slog.Duration("interval", w.interval),
	)

	// Fetch rates immediately on startup
	w.fetchWithRetry(ctx)

	// Backfill missing historical dates
	w.backfill(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("worker stopped due to context cancellation")
			return
		case <-w.stopCh:
			w.logger.Info("worker stopped")
			return
		case <-ticker.C:
			w.fetchWithRetry(ctx)
		}
	}
}

// Stop stops the worker.
func (w *Worker) Stop() {
	close(w.stopCh)
}

// FetchNow triggers an immediate fetch.
func (w *Worker) FetchNow(ctx context.Context) error {
	return w.fetch(ctx)
}

func (w *Worker) fetchWithRetry(ctx context.Context) {
	var lastErr error

	for attempt := 1; attempt <= w.retryAttempts; attempt++ {
		err := w.fetch(ctx)
		if err == nil {
			return
		}

		lastErr = err
		w.logger.Warn("fetch attempt failed",
			slog.Int("attempt", attempt),
			slog.Int("max_attempts", w.retryAttempts),
			slog.String("error", err.Error()),
		)

		if attempt < w.retryAttempts {
			select {
			case <-ctx.Done():
				return
			case <-time.After(w.retryDelay):
			}
		}
	}

	w.logger.Error("all fetch attempts failed",
		slog.String("error", lastErr.Error()),
	)
}

func (w *Worker) fetch(ctx context.Context) error {
	w.logger.Info("fetching currency rates")

	start := time.Now()

	rates, err := w.fetcher.FetchAllRates(ctx)
	if err != nil {
		metrics.WorkerFetchTotal.WithLabelValues("error").Inc()
		metrics.WorkerFetchDuration.Observe(time.Since(start).Seconds())
		return err
	}

	if err := w.repo.SaveBatch(ctx, rates); err != nil {
		metrics.WorkerFetchTotal.WithLabelValues("error").Inc()
		metrics.WorkerFetchDuration.Observe(time.Since(start).Seconds())
		return err
	}

	metrics.WorkerFetchTotal.WithLabelValues("success").Inc()
	metrics.WorkerFetchDuration.Observe(time.Since(start).Seconds())

	w.logger.Info("currency rates fetched and saved",
		slog.Int("count", len(rates)),
	)

	return nil
}

func (w *Worker) backfill(ctx context.Context) {
	if w.backfillDays <= 0 || w.historicalAPIURL == "" {
		return
	}

	now := time.Now().UTC().Truncate(24 * time.Hour)
	from := now.AddDate(0, 0, -w.backfillDays)
	to := now.AddDate(0, 0, -1) // yesterday

	w.logger.Info("starting backfill check",
		slog.String("from", from.Format("2006-01-02")),
		slog.String("to", to.Format("2006-01-02")),
	)

	existingDates, err := w.repo.GetExistingDates(ctx, from, to)
	if err != nil {
		w.logger.Error("backfill: failed to get existing dates", slog.String("error", err.Error()))
		return
	}

	existing := make(map[string]struct{}, len(existingDates))
	for _, d := range existingDates {
		existing[d.Format("2006-01-02")] = struct{}{}
	}

	var missing []time.Time
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		if _, ok := existing[d.Format("2006-01-02")]; !ok {
			missing = append(missing, d)
		}
	}

	if len(missing) == 0 {
		w.logger.Info("backfill: no gaps found")
		return
	}

	w.logger.Info("backfill: found missing dates", slog.Int("count", len(missing)))

	for _, date := range missing {
		select {
		case <-ctx.Done():
			w.logger.Info("backfill: interrupted by context cancellation")
			return
		default:
		}

		dateStr := date.Format("2006-01-02")

		rates, err := w.fetcher.FetchHistoricalRates(ctx, w.historicalAPIURL, date)
		if err != nil {
			metrics.WorkerBackfillTotal.WithLabelValues("error").Inc()
			w.logger.Warn("backfill: failed to fetch rates",
				slog.String("date", dateStr),
				slog.String("error", err.Error()),
			)
			continue
		}

		if err := w.repo.SaveBatch(ctx, rates); err != nil {
			metrics.WorkerBackfillTotal.WithLabelValues("error").Inc()
			w.logger.Warn("backfill: failed to save rates",
				slog.String("date", dateStr),
				slog.String("error", err.Error()),
			)
			continue
		}

		metrics.WorkerBackfillTotal.WithLabelValues("success").Inc()
		w.logger.Info("backfill: saved rates",
			slog.String("date", dateStr),
			slog.Int("count", len(rates)),
		)

		time.Sleep(500 * time.Millisecond)
	}

	w.logger.Info("backfill: completed")
}
