package app

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func LoadConfig() Config {
	return Config{
		Port:        getEnv("PORT", "8081"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/payments_db?sslmode=disable"),
	}
}

// NewDB opens and verifies a PostgreSQL connection.
func NewDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	log.Println("[payment-service] database connection established")
	return db, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
