package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	orderv1 "github.com/usenbai-nur/ADP2_asik2_generated/order/v1"
	paymentv1 "github.com/usenbai-nur/ADP2_asik2_generated/payment/v1"

	"github.com/ap2/order-service/internal/app"
	"github.com/ap2/order-service/internal/repository"
	transportGRPC "github.com/ap2/order-service/internal/transport/grpc"
	transportHTTP "github.com/ap2/order-service/internal/transport/http"
	"github.com/ap2/order-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := app.LoadConfig()

	db, err := app.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[order-service] failed to connect to database: %v", err)
	}
	defer db.Close()

	grpcConn, err := grpc.NewClient(
		cfg.PaymentGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("[order-service] failed to create payment grpc client: %v", err)
	}
	defer grpcConn.Close()

	orderRepo := repository.NewPostgresOrderRepository(db)
	paymentClient := repository.NewGRPCPaymentClient(
		paymentv1.NewPaymentServiceClient(grpcConn),
		cfg.PaymentCallTimeout,
	)

	statusHub := app.NewOrderStatusHub()

	orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient, statusHub)

	orderHandler := transportHTTP.NewOrderHandler(orderUC)
	orderGRPCServer := transportGRPC.NewOrderServer(orderRepo, statusHub)

	httpRouter := gin.Default()
	httpRouter.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "order-service",
			"status":  "ok",
		})
	})
	orderHandler.RegisterRoutes(httpRouter)

	httpServer := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           httpRouter,
		ReadHeaderTimeout: 5 * time.Second,
	}

	grpcServer := grpc.NewServer()
	orderv1.RegisterOrderServiceServer(grpcServer, orderGRPCServer)

	grpcListener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("[order-service] failed to listen grpc: %v", err)
	}

	httpErrCh := make(chan error, 1)
	grpcErrCh := make(chan error, 1)

	go func() {
		log.Printf("[order-service] http server starting on :%s", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			httpErrCh <- err
		}
	}()

	go func() {
		log.Printf("[order-service] grpc server starting on :%s", cfg.GRPCPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			grpcErrCh <- err
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		log.Println("[order-service] shutdown signal received")
	case err := <-httpErrCh:
		log.Fatalf("[order-service] http server error: %v", err)
	case err := <-grpcErrCh:
		log.Fatalf("[order-service] grpc server error: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcServer.GracefulStop()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("[order-service] http shutdown error: %v", err)
	}

	log.Println("[order-service] stopped")
}