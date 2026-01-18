package worker

import (
	"context"
	"log/slog"
	"time"
)

// Worker handles periodic currency rate fetching.
type Worker struct {
	fetcher       *Fetcher
	repo          RateRepository
	logger        *slog.Logger
	interval      time.Duration
	retryAttempts int
	retryDelay    time.Duration
	stopCh        chan struct{}
}

// NewWorker creates a new worker instance.
func NewWorker(
	fetcher *Fetcher,
	repo RateRepository,
	logger *slog.Logger,
	interval time.Duration,
	retryAttempts int,
	retryDelay time.Duration,
) *Worker {
	return &Worker{
		fetcher:       fetcher,
		repo:          repo,
		logger:        logger,
		interval:      interval,
		retryAttempts: retryAttempts,
		retryDelay:    retryDelay,
		stopCh:        make(chan struct{}),
	}
}

// Start begins the periodic rate fetching.
func (w *Worker) Start(ctx context.Context) {
	w.logger.Info("starting currency rate worker",
		slog.Duration("interval", w.interval),
	)

	// Fetch rates immediately on startup
	w.fetchWithRetry(ctx)

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

	rates, err := w.fetcher.FetchAllRates(ctx)
	if err != nil {
		return err
	}

	if err := w.repo.SaveBatch(ctx, rates); err != nil {
		return err
	}

	w.logger.Info("currency rates fetched and saved",
		slog.Int("count", len(rates)),
	)

	return nil
}
