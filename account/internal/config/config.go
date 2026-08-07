package config

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

// Config — конфигурация сервиса из переменных окружения.
type Config struct {
	ServiceName  string `env:"SERVICE_NAME" envDefault:"account"`
	AppEnv       string `env:"APP_ENV" envDefault:"local"`
	LogLevel     string `env:"LOG_LEVEL" envDefault:"debug"`
	HTTPPort     int    `env:"HTTP_PORT" envDefault:"8082"`
	GRPCPort     int    `env:"GRPC_PORT" envDefault:"50052"`
	DBDSN        string `env:"DB_DSN" envDefault:"postgres://ledger:ledger@localhost:5432/ledgerpay?sslmode=disable"`
	RedisAddr    string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	KafkaBrokers string `env:"KAFKA_BROKERS" envDefault:"localhost:19092"`
	JWTSecret    string `env:"JWT_SECRET" envDefault:"dev-change-me-please-32chars-min"`
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
