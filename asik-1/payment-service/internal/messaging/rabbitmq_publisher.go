package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ap2/payment-service/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

const PaymentCompletedQueue = "payment.completed"

type RabbitMQPublisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbitMQPublisher(url string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		PaymentCompletedQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	return &RabbitMQPublisher{
		conn: conn,
		ch:   ch,
	}, nil
}

func (p *RabbitMQPublisher) PublishPaymentCompleted(ctx context.Context, event domain.PaymentCompletedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal payment event: %w", err)
	}

	err = p.ch.PublishWithContext(
		ctx,
		"",
		PaymentCompletedQueue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish payment event: %w", err)
	}

	return nil
}

func (p *RabbitMQPublisher) Close() error {
	if p.ch != nil {
		_ = p.ch.Close()
	}

	if p.conn != nil {
		return p.conn.Close()
	}

	return nil
}