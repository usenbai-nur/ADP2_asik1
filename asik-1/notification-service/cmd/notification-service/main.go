package main

import (
	"context"
	"log"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ap2/notification-service/internal/app"
	"github.com/ap2/notification-service/internal/domain"
	"github.com/ap2/notification-service/internal/messaging"
	"github.com/ap2/notification-service/internal/provider"
	"github.com/ap2/notification-service/internal/repository"
)

func main() {
	cfg := app.LoadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var sender domain.EmailSender

	switch strings.ToUpper(cfg.ProviderMode) {
	case "SIMULATED":
		sender = provider.NewSimulatedEmailSender()
	default:
		log.Fatalf("[notification-service] unsupported PROVIDER_MODE=%s", cfg.ProviderMode)
	}

	idempotencyStore := repository.NewRedisIdempotencyStore(cfg.RedisAddr, cfg.IdempotencyTTL)
	defer idempotencyStore.Close()

	consumer, err := messaging.NewRabbitMQConsumer(
		cfg.RabbitMQURL,
		cfg.QueueName,
		sender,
		idempotencyStore,
		cfg.MaxRetries,
	)
	if err != nil {
		log.Fatalf("[notification-service] failed to connect rabbitmq: %v", err)
	}
	defer consumer.Close()

	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("[notification-service] consumer error: %v", err)
	}

	log.Println("[notification-service] stopped gracefully")
}