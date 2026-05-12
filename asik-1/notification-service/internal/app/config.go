package app

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	RabbitMQURL    string
	QueueName      string
	RedisAddr      string
	ProviderMode   string
	MaxRetries     int
	IdempotencyTTL time.Duration
}

func LoadConfig() Config {
	return Config{
		RabbitMQURL:    getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		QueueName:      getEnv("PAYMENT_COMPLETED_QUEUE", "payment.completed"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		ProviderMode:   getEnv("PROVIDER_MODE", "SIMULATED"),
		MaxRetries:     getEnvInt("MAX_RETRIES", 3),
		IdempotencyTTL: getEnvDuration("IDEMPOTENCY_TTL", 24*time.Hour),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}