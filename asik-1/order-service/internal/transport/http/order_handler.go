package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/ap2/order-service/internal/domain"
	"github.com/ap2/order-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderUseCase *usecase.OrderUseCase
}

// NewOrderHandler wires the handler to its use case.
func NewOrderHandler(uc *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{orderUseCase: uc}
}

// RegisterRoutes attaches all order routes to the given Gin engine.
func (h *OrderHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/orders", h.CreateOrder)
	r.GET("/orders/stats", h.GetOrderStats)
	r.GET("/orders/:id", h.GetOrder)
	r.PATCH("/orders/:id/cancel", h.CancelOrder)
}


// IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII
// GetOrderStats handles GET /orders/stats
func (h *OrderHandler) GetOrderStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	stats, err := h.orderUseCase.GetOrderStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch stats"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// Request / Response DTOs

type createOrderRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
	ItemName   string `json:"item_name"   binding:"required"`
	Amount     int64  `json:"amount"      binding:"required,gt=0"`
}

type orderResponse struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	ItemName   string    `json:"item_name"`
	Amount     int64     `json:"amount"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func toOrderResponse(o *domain.Order) orderResponse {
	return orderResponse{
		ID:         o.ID,
		CustomerID: o.CustomerID,
		ItemName:   o.ItemName,
		Amount:     o.Amount,
		Status:     o.Status,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,
	}
}

// Handlers

// CreateOrder handles POST /orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	idempotencyKey := c.GetHeader("Idempotency-Key") // bonus

	output, err := h.orderUseCase.CreateOrder(ctx, usecase.CreateOrderInput{
		CustomerID:     req.CustomerID,
		ItemName:       req.ItemName,
		Amount:         req.Amount,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		if errors.Is(err, domain.ErrPaymentUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "payment service unavailable, order marked as failed"})
			return
		}
		if errors.Is(err, domain.ErrInvalidAmount) ||
			errors.Is(err, domain.ErrInvalidCustomerID) ||
			errors.Is(err, domain.ErrInvalidItemName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	status := http.StatusCreated
	if output.Order.Status == domain.StatusFailed {
		status = http.StatusPaymentRequired
	}
	c.JSON(status, toOrderResponse(output.Order))
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.orderUseCase.GetOrder(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, toOrderResponse(order))
}

// CancelOrder handles PATCH /orders/:id/cancel
// If the order is already paid or failed, it cannot be cancelled and a 409 Conflict is returned.
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.orderUseCase.CancelOrder(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		if errors.Is(err, domain.ErrCannotCancel) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, toOrderResponse(order))
}
