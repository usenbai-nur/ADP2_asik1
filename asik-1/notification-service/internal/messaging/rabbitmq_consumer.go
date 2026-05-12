package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ap2/notification-service/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

type IdempotencyStore interface {
	IsProcessed(ctx context.Context, eventID string) (bool, error)
	MarkProcessed(ctx context.Context, eventID string) error
}

type RabbitMQConsumer struct {
	conn       *amqp.Connection
	ch         *amqp.Channel
	queueName  string
	sender     domain.EmailSender
	store      IdempotencyStore
	maxRetries int
}

func NewRabbitMQConsumer(
	url string,
	queueName string,
	sender domain.EmailSender,
	store IdempotencyStore,
	maxRetries int,
) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("set qos: %w", err)
	}

	return &RabbitMQConsumer{
		conn:       conn,
		ch:         ch,
		queueName:  queueName,
		sender:     sender,
		store:      store,
		maxRetries: maxRetries,
	}, nil
}

func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	messages, err := c.ch.Consume(
		c.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume messages: %w", err)
	}

	log.Printf("[notification-service] listening queue: %s", c.queueName)

	for {
		select {
		case <-ctx.Done():
			return nil

		case msg, ok := <-messages:
			if !ok {
				return nil
			}

			if err := c.handleMessage(ctx, msg); err != nil {
				log.Printf("[notification-service] failed to process message: %v", err)

				if nackErr := msg.Nack(false, true); nackErr != nil {
					log.Printf("[notification-service] failed to nack message: %v", nackErr)
				}

				continue
			}

			if ackErr := msg.Ack(false); ackErr != nil {
				log.Printf("[notification-service] failed to ack message: %v", ackErr)
			}
		}
	}
}

func (c *RabbitMQConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) error {
	var event domain.PaymentCompletedEvent

	if err := json.Unmarshal(msg.Body, &event); err != nil {
		return err
	}

	if event.EventID == "" {
		return fmt.Errorf("event_id is empty")
	}

	processed, err := c.store.IsProcessed(ctx, event.EventID)
	if err != nil {
		return err
	}

	if processed {
		log.Printf("[notification-service] duplicate event ignored: %s", event.EventID)
		return nil
	}

	var lastErr error

	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		lastErr = c.sender.SendPaymentNotification(ctx, event)
		if lastErr == nil {
			if err := c.store.MarkProcessed(ctx, event.EventID); err != nil {
				return err
			}

			log.Printf("[notification-service] notification processed successfully event_id=%s", event.EventID)
			return nil
		}

		backoff := time.Duration(1<<attempt) * time.Second

		log.Printf(
			"[notification-service] provider failed attempt=%d/%d error=%v retry_in=%s",
			attempt,
			c.maxRetries,
			lastErr,
			backoff,
		)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
	}

	return fmt.Errorf("notification failed after retries: %w", lastErr)
}

func (c *RabbitMQConsumer) Close() error {
	if c.ch != nil {
		_ = c.ch.Close()
	}

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}