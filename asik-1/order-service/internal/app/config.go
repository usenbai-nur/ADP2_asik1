package app

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	HTTPPort           string
	GRPCPort           string
	DatabaseURL        string
	PaymentGRPCAddr    string
	PaymentCallTimeout time.Duration
}

func LoadConfig() Config {
	return Config{
		HTTPPort:           getEnv("PORT", "8080"),
		GRPCPort:           getEnv("ORDER_GRPC_PORT", "50052"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/orders_db?sslmode=disable"),
		PaymentGRPCAddr:    getEnv("PAYMENT_GRPC_ADDR", "localhost:50051"),
		PaymentCallTimeout: getEnvDuration("PAYMENT_CALL_TIMEOUT", 2*time.Second),
	}
}

func NewDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	log.Println("[order-service] database connection established")
	return db, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return parsed
}