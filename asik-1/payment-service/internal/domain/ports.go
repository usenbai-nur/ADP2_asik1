package domain

import (
	"context"
	"errors"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrInvalidOrderID  = errors.New("order_id is required")
	ErrInvalidAmount   = errors.New("amount must be greater than zero")
)

type PaymentRepository interface {
	Save(ctx context.Context, payment *Payment) error
	FindByOrderID(ctx context.Context, orderID string) (*Payment, error)
}
