package domain

import "context"

// OrderRepository defines persistence operations for orders.
type OrderRepository interface {
	Save(ctx context.Context, order *Order) error
	FindByID(ctx context.Context, id string) (*Order, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*Order, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	CountByStatus(ctx context.Context) (map[string]int, error)
}

type PaymentResult struct {
	TransactionID string
	Status        string // "Authorized" | "Declined"
}

type PaymentClient interface {
	Authorize(ctx context.Context, orderID string, amount int64) (*PaymentResult, error)
}