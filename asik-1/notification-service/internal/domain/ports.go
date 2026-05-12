package domain

import "context"

type EmailSender interface {
	SendPaymentNotification(ctx context.Context, event PaymentCompletedEvent) error
}