package domain

import (
	"context"
	"time"
)

type OrderRepository interface {
	Save(ctx context.Context, order *Order) error
	FindByID(ctx context.Context, id string) (*Order, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*Order, error)
	UpdateStatus(ctx context.Context, id string, status string, updatedAt time.Time) error
	CountByStatus(ctx context.Context) (map[string]int, error)
}

type PaymentResult struct {
	TransactionID string
	Status        string
}

type PaymentClient interface {
	Authorize(ctx context.Context, orderID string, amount int64) (*PaymentResult, error)
}

type OrderStatusPublisher interface {
	Publish(event OrderStatusEvent)
}