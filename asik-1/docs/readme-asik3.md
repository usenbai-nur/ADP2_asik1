# Event-Driven Microservices System (Order–Payment–Notification)

## Overview

This project implements a microservices architecture using:

- Go (Golang)
- gRPC (service-to-service communication)
- RabbitMQ (event-driven messaging)
- PostgreSQL (data storage)
- Docker Compose (environment orchestration)

The system consists of three services:

- **Order Service** – handles HTTP requests and manages orders
- **Payment Service** – processes payments via gRPC
- **Notification Service** – consumes events and sends notifications

---

## Architecture Flow

1. Client sends HTTP request to Order Service
2. Order Service calls Payment Service via gRPC
3. Payment Service processes payment
4. Payment Service publishes event to RabbitMQ
5. Notification Service consumes event
6. Notification is sent (logged)

---

## Idempotency Strategy

Idempotency is implemented in the **Notification Service**.

Each event includes a unique:

event_id (UUID)


### How it works:

- Notification Service stores processed event IDs in memory (`map[string]bool`)
- When a message is received:
  - If `event_id` already exists → message is ignored
  - If not → message is processed and stored

### Result:

- Prevents duplicate notifications
- Handles message re-delivery safely

---

## ACK Logic

Manual acknowledgment (ACK) is used in RabbitMQ consumer.

### Flow:

1. Message is received from queue
2. Event is processed
3. If successful → `Ack()` is called
4. If failed → message is not acknowledged

### Why this matters:

- Prevents message loss
- Ensures reliability
- Allows re-delivery in case of failure

---

## Persistent Messaging

The system uses:

- **Durable Queue**
- **Persistent Messages**

```go
DeliveryMode: amqp.Persistent

This improves reliability and ensures messages are not lost during broker restarts.
```

## Running the Project

```
docker compose up --build
```

## Testing
### Create Order

- POST http://localhost:8080/orders

Body:
```
{
  "customer_id": "cust-31",
  "item_name": "banan",
  "amount": 300
}
```
Expected result:
- Order created
- Payment processed
- Event published
- Notification logged