package provider

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/ap2/notification-service/internal/domain"
)

type SimulatedEmailSender struct {
	failurePercent int
	latency        time.Duration
}

func NewSimulatedEmailSender() *SimulatedEmailSender {
	return &SimulatedEmailSender{
		failurePercent: 30,
		latency:        1 * time.Second,
	}
}

func (s *SimulatedEmailSender) SendPaymentNotification(ctx context.Context, event domain.PaymentCompletedEvent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(s.latency):
	}

	if rand.Intn(100) < s.failurePercent {
		return fmt.Errorf("simulated email provider temporary failure")
	}

	log.Printf(
		"[Provider:SIMULATED] Email sent to %s for Order #%s. Amount: $%.2f. Status: %s",
		event.CustomerEmail,
		event.OrderID,
		float64(event.Amount)/100,
		event.Status,
	)

	return nil
}