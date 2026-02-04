package config

import (
	"time"

	"github.com/caarlos0/env/v10"
)

// Config holds the gateway service configuration.
type Config struct {
	Server   ServerConfig
	JWT      JWTConfig
	Currency CurrencyConfig
	Log      LogConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port         int           `env:"GATEWAY_PORT" envDefault:"8080"`
	Host         string        `env:"GATEWAY_HOST" envDefault:"0.0.0.0"`
	ReadTimeout  time.Duration `env:"GATEWAY_READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout time.Duration `env:"GATEWAY_WRITE_TIMEOUT" envDefault:"15s"`
}

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	Secret     string        `env:"JWT_SECRET" envDefault:"your-super-secret-key-change-in-production-32"`
	Expiration time.Duration `env:"JWT_EXPIRATION" envDefault:"24h"`
	Issuer     string        `env:"JWT_ISSUER" envDefault:"currency-gateway"`
}

// CurrencyConfig holds currency service connection configuration.
type CurrencyConfig struct {
	ServiceURL string        `env:"CURRENCY_SERVICE_URL" envDefault:"http://localhost:8081"`
	Timeout    time.Duration `env:"CURRENCY_SERVICE_TIMEOUT" envDefault:"10s"`
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
