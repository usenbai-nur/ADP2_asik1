package domain

import "context"

type PaymentRepository interface {
	Save(ctx context.Context, payment *Payment) error
	FindByOrderID(ctx context.Context, orderID string) (*Payment, error)
	GetStats(ctx context.Context) (*PaymentStats, error)
}

type PaymentEventPublisher interface {
	PublishPaymentCompleted(ctx context.Context, event PaymentCompletedEvent) error
	Close() error
}