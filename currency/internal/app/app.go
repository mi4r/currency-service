package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mi4r/currency-service/currency/internal/config"
	"github.com/mi4r/currency-service/currency/internal/handler"
	"github.com/mi4r/currency-service/currency/internal/repository"
	"github.com/mi4r/currency-service/currency/internal/service"
	"github.com/mi4r/currency-service/currency/internal/worker"
	"github.com/mi4r/currency-service/pkg/logger"
)

// App represents the application.
type App struct {
	cfg    *config.Config
	logger *slog.Logger
	server *http.Server
	worker *worker.Worker
	pool   *pgxpool.Pool
}

// New creates a new application instance.
func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	log := logger.New(cfg.Log.Level)

	return &App{
		cfg:    cfg,
		logger: log,
	}, nil
}

// Run starts the application.
func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to database
	pool, err := a.connectDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	a.pool = pool
	defer pool.Close()

	// Initialize dependencies
	repo := repository.NewPostgresRepository(pool)
	svc := service.NewCurrencyService(repo, a.logger)
	h := handler.NewHandler(svc, a.logger)

	// Setup router
	router := chi.NewRouter()
	handler.RegisterRoutes(router, h)

	// Setup HTTP server
	a.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.cfg.Server.Host, a.cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  a.cfg.Server.ReadTimeout,
		WriteTimeout: a.cfg.Server.WriteTimeout,
	}

	// Setup worker
	fetcher := worker.NewFetcher(a.cfg.Worker.APIURL)
	a.worker = worker.NewWorker(
		fetcher,
		repo,
		a.logger,
		a.cfg.Worker.FetchInterval,
		a.cfg.Worker.RetryAttempts,
		a.cfg.Worker.RetryDelay,
		a.cfg.Worker.HistoricalAPIURL,
		a.cfg.Worker.BackfillDays,
	)

	// Start worker in background
	go a.worker.Start(ctx)

	// Start HTTP server
	go func() {
		a.logger.Info("starting HTTP server",
			slog.String("addr", a.server.Addr),
		)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("HTTP server error", slog.String("error", err.Error()))
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("server shutdown error", slog.String("error", err.Error()))
	}

	a.worker.Stop()
	cancel()

	a.logger.Info("server stopped")
	return nil
}

func (a *App) connectDB(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		a.cfg.Database.User,
		a.cfg.Database.Password,
		a.cfg.Database.Host,
		a.cfg.Database.Port,
		a.cfg.Database.Name,
		a.cfg.Database.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	poolConfig.MaxConns = int32(a.cfg.Database.MaxOpenConns)
	poolConfig.MinConns = int32(a.cfg.Database.MaxIdleConns)
	poolConfig.MaxConnLifetime = a.cfg.Database.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	a.logger.Info("connected to database",
		slog.String("host", a.cfg.Database.Host),
		slog.Int("port", a.cfg.Database.Port),
	)

	return pool, nil
}
