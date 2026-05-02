package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"
	"time"

	"github.com/ap2/payment-service/internal/app"
	"github.com/ap2/payment-service/internal/messaging"
	"github.com/ap2/payment-service/internal/repository"
	transportGRPC "github.com/ap2/payment-service/internal/transport/grpc"
	"github.com/ap2/payment-service/internal/usecase"
	paymentv1 "github.com/usenbai-nur/ADP2_asik2_generated/payment/v1"
	grpcpkg "google.golang.org/grpc"
)

func main() {
	cfg := app.LoadConfig()

	db, err := app.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[payment-service] failed to connect to database: %v", err)
	}
	defer db.Close()

	paymentRepo := repository.NewPostgresPaymentRepository(db)

	publisher, err := messaging.NewRabbitMQPublisher(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("[payment-service] failed to connect rabbitmq: %v", err)
	}
	defer publisher.Close()

	paymentUC := usecase.NewPaymentUseCase(paymentRepo, publisher)

	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("[payment-service] failed to listen grpc: %v", err)
	}

	grpcServer := grpcpkg.NewServer(
		grpcpkg.UnaryInterceptor(transportGRPC.LoggingInterceptor()),
	)

	paymentServer := transportGRPC.NewPaymentServer(paymentUC)
	paymentv1.RegisterPaymentServiceServer(grpcServer, paymentServer)

	go func() {
		log.Printf("[payment-service] grpc server starting on :%s", cfg.GRPCPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("[payment-service] grpc server error: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("[payment-service] shutdown signal received")

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Println("[payment-service] stopped")
	case <-time.After(10 * time.Second):
		log.Println("[payment-service] forcing stop")
		grpcServer.Stop()
	}
}