package main

import (
	"log"
	"net/http"

	"github.com/ap2/payment-service/internal/app"
	"github.com/ap2/payment-service/internal/repository"
	transportHTTP "github.com/ap2/payment-service/internal/transport/http"
	"github.com/ap2/payment-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := app.LoadConfig()

	// Database 
	db, err := app.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[payment-service] failed to connect to database: %v", err)
	}
	defer db.Close()
	// Dependency injection 
	paymentRepo := repository.NewPostgresPaymentRepository(db)

	paymentUC := usecase.NewPaymentUseCase(paymentRepo)

	paymentHandler := transportHTTP.NewPaymentHandler(paymentUC)

	// Router
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "payment-service", "status": "ok"})
	})
	paymentHandler.RegisterRoutes(router)

	log.Printf("[payment-service] starting on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("[payment-service] server error: %v", err)
	}
}
