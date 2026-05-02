package app

import "os"

type Config struct {
	RabbitMQURL string
	QueueName   string
}

func LoadConfig() Config {
	return Config{
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		QueueName:   getEnv("PAYMENT_COMPLETED_QUEUE", "payment.completed"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}