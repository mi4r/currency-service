package config

import (
	"time"

	"github.com/caarlos0/env/v10"
)

// Config holds the currency service configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Worker   WorkerConfig
	Log      LogConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port         int           `env:"CURRENCY_PORT" envDefault:"8081"`
	Host         string        `env:"CURRENCY_HOST" envDefault:"0.0.0.0"`
	ReadTimeout  time.Duration `env:"CURRENCY_READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout time.Duration `env:"CURRENCY_WRITE_TIMEOUT" envDefault:"15s"`
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host            string        `env:"DB_HOST" envDefault:"localhost"`
	Port            int           `env:"DB_PORT" envDefault:"5432"`
	User            string        `env:"DB_USER" envDefault:"currency_user"`
	Password        string        `env:"DB_PASSWORD" envDefault:"currency_password"`
	Name            string        `env:"DB_NAME" envDefault:"currency_db"`
	SSLMode         string        `env:"DB_SSLMODE" envDefault:"disable"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"5m"`
}

// WorkerConfig holds worker configuration.
type WorkerConfig struct {
	FetchInterval  time.Duration `env:"WORKER_FETCH_INTERVAL" envDefault:"24h"`
	APIURL         string        `env:"WORKER_API_URL" envDefault:"https://latest.currency-api.pages.dev/v1/currencies/rub.json"`
	TargetCurrency string        `env:"WORKER_TARGET_CURRENCY" envDefault:"usd"`
	RetryAttempts  int           `env:"WORKER_RETRY_ATTEMPTS" envDefault:"3"`
	RetryDelay     time.Duration `env:"WORKER_RETRY_DELAY" envDefault:"5m"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string `env:"LOG_LEVEL" envDefault:"info"`
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
