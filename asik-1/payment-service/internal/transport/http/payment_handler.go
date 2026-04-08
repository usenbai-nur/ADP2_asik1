package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/ap2/payment-service/internal/domain"
	"github.com/ap2/payment-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentUseCase *usecase.PaymentUseCase
}

func NewPaymentHandler(uc *usecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{paymentUseCase: uc}
}

func (h *PaymentHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/payments", h.Authorize)
	r.GET("/payments/:order_id", h.GetByOrderID)
}

// DTOs

type authorizeRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	Amount  int64  `json:"amount"   binding:"required,gt=0"`
}

type paymentResponse struct {
	ID            string    `json:"id"`
	OrderID       string    `json:"order_id"`
	TransactionID string    `json:"transaction_id"`
	Amount        int64     `json:"amount"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

func toPaymentResponse(p *domain.Payment) paymentResponse {
	return paymentResponse{
		ID:            p.ID,
		OrderID:       p.OrderID,
		TransactionID: p.TransactionID,
		Amount:        p.Amount,
		Status:        p.Status,
		CreatedAt:     p.CreatedAt,
	}
}

// Handlers

// Authorize handles POST /payments
func (h *PaymentHandler) Authorize(c *gin.Context) {
	var req authorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := h.paymentUseCase.Authorize(c.Request.Context(), usecase.AuthorizeInput{
		OrderID: req.OrderID,
		Amount:  req.Amount,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidOrderID) || errors.Is(err, domain.ErrInvalidAmount) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, toPaymentResponse(output.Payment))
}

// GetByOrderID handles GET /payments/:order_id
func (h *PaymentHandler) GetByOrderID(c *gin.Context) {
	orderID := c.Param("order_id")
	payment, err := h.paymentUseCase.GetByOrderID(c.Request.Context(), orderID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, toPaymentResponse(payment))
}
