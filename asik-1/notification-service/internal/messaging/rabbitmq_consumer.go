package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/ap2/notification-service/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConsumer struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	queueName string

	mu        sync.Mutex
	processed map[string]bool
}

func NewRabbitMQConsumer(url string, queueName string) (*RabbitMQConsumer, error) {
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
		conn:      conn,
		ch:        ch,
		queueName: queueName,
		processed: make(map[string]bool),
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

			if err := c.handleMessage(msg); err != nil {
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

func (c *RabbitMQConsumer) handleMessage(msg amqp.Delivery) error {
	var event domain.PaymentCompletedEvent

	if err := json.Unmarshal(msg.Body, &event); err != nil {
		return err
	}

	if event.EventID == "" {
		return fmt.Errorf("event_id is empty")
	}

	c.mu.Lock()
	if c.processed[event.EventID] {
		c.mu.Unlock()
		log.Printf("[notification-service] duplicate event ignored: %s", event.EventID)
		return nil
	}

	c.processed[event.EventID] = true
	c.mu.Unlock()

	log.Printf(
		"[Notification] Sent email to %s for Order #%s. Amount: $%.2f. Status: %s",
		event.CustomerEmail,
		event.OrderID,
		float64(event.Amount)/100,
		event.Status,
	)

	return nil
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