package main

import (
	"log"
	"net/http"
	"time"

	"github.com/ap2/order-service/internal/app"
	"github.com/ap2/order-service/internal/repository"
	transportHTTP "github.com/ap2/order-service/internal/transport/http"
	"github.com/ap2/order-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := app.LoadConfig()

	// Database 
	db, err := app.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[order-service] failed to connect to database: %v", err)
	}
	defer db.Close()

	httpClient := &http.Client{
		Timeout: 2 * time.Second,
	}

	orderRepo := repository.NewPostgresOrderRepository(db)
	paymentClient := repository.NewHTTPPaymentClient(cfg.PaymentBaseURL, httpClient)

	orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient)

	orderHandler := transportHTTP.NewOrderHandler(orderUC)

	// Router
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "order-service", "status": "ok"})
	})
	orderHandler.RegisterRoutes(router)

	log.Printf("[order-service] starting on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("[order-service] server error: %v", err)
	}
}
