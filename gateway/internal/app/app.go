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

	"github.com/mi4r/currency-service/gateway/internal/client"
	"github.com/mi4r/currency-service/gateway/internal/config"
	"github.com/mi4r/currency-service/gateway/internal/handler"
	"github.com/mi4r/currency-service/gateway/internal/repository"
	"github.com/mi4r/currency-service/gateway/internal/service"
	"github.com/mi4r/currency-service/pkg/jwt"
	"github.com/mi4r/currency-service/pkg/logger"
)

// App represents the gateway application.
type App struct {
	cfg    *config.Config
	logger *slog.Logger
	server *http.Server
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
	// Initialize JWT manager
	jwtManager := jwt.NewManager(
		a.cfg.JWT.Secret,
		a.cfg.JWT.Expiration,
		a.cfg.JWT.Issuer,
	)

	// Initialize repositories
	userRepo := repository.NewMemoryUserRepository()

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtManager, a.logger)

	// Initialize clients
	currencyClient := client.NewCurrencyClient(
		a.cfg.Currency.ServiceURL,
		a.cfg.Currency.Timeout,
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, a.logger)
	currencyHandler := handler.NewCurrencyHandler(currencyClient, a.logger)

	// Setup router
	router := chi.NewRouter()
	handler.RegisterRoutes(router, authHandler, currencyHandler, jwtManager, a.logger)

	// Setup HTTP server
	a.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.cfg.Server.Host, a.cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  a.cfg.Server.ReadTimeout,
		WriteTimeout: a.cfg.Server.WriteTimeout,
	}

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

	a.logger.Info("server stopped")
	return nil
}
