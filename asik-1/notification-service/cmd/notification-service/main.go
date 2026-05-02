package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/ap2/notification-service/internal/app"
	"github.com/ap2/notification-service/internal/messaging"
)

func main() {
	cfg := app.LoadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	consumer, err := messaging.NewRabbitMQConsumer(cfg.RabbitMQURL, cfg.QueueName)
	if err != nil {
		log.Fatalf("[notification-service] failed to connect rabbitmq: %v", err)
	}
	defer consumer.Close()

	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("[notification-service] consumer error: %v", err)
	}

	log.Println("[notification-service] stopped gracefully")
}