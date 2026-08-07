package config

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

// Config — конфигурация gateway из переменных окружения.
type Config struct {
	ServiceName           string `env:"SERVICE_NAME" envDefault:"gateway"`
	AppEnv                string `env:"APP_ENV" envDefault:"local"`
	LogLevel              string `env:"LOG_LEVEL" envDefault:"debug"`
	HTTPPort              int    `env:"HTTP_PORT" envDefault:"8080"`
	JWTSecret             string `env:"JWT_SECRET" envDefault:"dev-change-me-please-32chars-min"`
	AuthGRPCAddr          string `env:"AUTH_GRPC_ADDR" envDefault:"localhost:50051"`
	AccountGRPCAddr       string `env:"ACCOUNT_GRPC_ADDR" envDefault:"localhost:50052"`
	TransactionGRPCAddr   string `env:"TRANSACTION_GRPC_ADDR" envDefault:"localhost:50053"`
	NotificationsGRPCAddr string `env:"NOTIFICATIONS_GRPC_ADDR" envDefault:"localhost:50054"`
}

// Load читает .env (если есть) и парсит переменные окружения.
func Load() (*Config, error) {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("разбор env: %w", err)
	}
	return cfg, nil
}

// MustLoad загружает конфиг или завершает процесс при ошибке.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		log.Fatalf("конфиг: %v", err)
	}
	return cfg
}
