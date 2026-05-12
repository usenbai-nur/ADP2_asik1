# Assignment 4 — Performance Optimization & External Integrations



# Project Overview

This assignment extends the existing microservice architecture by adding:

- Redis cache-aside pattern
- Cache invalidation
- RabbitMQ background worker
- Retry mechanism with exponential backoff
- Redis idempotency
- Adapter pattern for external providers
- Dockerized infrastructure

The system is designed to be more scalable, reliable, fault tolerant, and production-oriented.

---

# Microservices

The project contains three main microservices:

## 1. Order Service

Responsibilities:

- Create orders
- Retrieve order information
- Cancel orders
- Communicate with Payment Service using gRPC
- Use Redis cache-aside strategy

Technologies:

- Go
- Gin
- gRPC
- PostgreSQL
- Redis

---

## 2. Payment Service

Responsibilities:

- Authorize payments
- Store payment transactions
- Publish payment events to RabbitMQ

Technologies:

- Go
- gRPC
- PostgreSQL
- RabbitMQ

---

## 3. Notification Service

Responsibilities:

- Consume RabbitMQ events
- Process notifications asynchronously
- Retry failed provider calls
- Prevent duplicate processing using Redis
- Use provider adapter abstraction

Technologies:

- Go
- RabbitMQ
- Redis

---

# Infrastructure Components

## PostgreSQL

Used as the primary persistent database.

Separate databases:

- orders_db
- payments_db

---

## RabbitMQ

RabbitMQ is used for asynchronous event-driven communication.

Queue used:

```text
payment.completed
```

Payment Service publishes events, and Notification Service consumes them.

Benefits:

- Loose coupling
- Scalability
- Background processing
- Fault tolerance

---

## Redis

Redis is used for:

- Cache-aside pattern
- Cache invalidation
- Idempotency storage

Benefits:

- Faster reads
- Reduced database load
- Duplicate prevention

---

# Redis Cache-Aside Pattern

Implemented inside Order Service.

Flow:

1. Client requests order data.
2. Order Service checks Redis first.
3. If cache exists → return cached object.
4. If cache does not exist:
   - load order from PostgreSQL
   - store order in Redis
   - return response

Redis key format:

```
order:{order_id}
```

TTL:

```
5 minutes
```

---

# Cache Invalidation

Whenever order status changes:

- Pending → Paid
- Pending → Failed
- Pending → Cancelled

Order Service invalidates Redis cache.

Flow:

```
Update PostgreSQL
→ Delete Redis key
→ Next request reloads fresh data
```

This prevents stale cache.

---

# Background Worker

Notification Service works as a background worker.

Flow:

```
RabbitMQ event
→ Notification Service consumes event
→ Sends notification asynchronously
```

The HTTP request does not wait for notification delivery.

Benefits:

- Faster API response
- Better scalability
- Non-blocking architecture

---

# Retry Mechanism

Notification sending may fail temporarily.

The system retries automatically.

Configuration:

```
MAX_RETRIES=3
```

Retry strategy:

```
2s → 4s → 8s
```

This is called exponential backoff.

Benefits:

- Improved reliability
- Better fault tolerance
- Temporary failure recovery

---

# Idempotency

Redis is used to prevent duplicate event processing.

Redis key format:

```
notification:event:{event_id}
```

Flow:

1. Worker receives event
2. Check Redis key
3. If event already processed → skip
4. Otherwise process event
5. Save processed state to Redis

Benefits:

- Prevent duplicate notifications
- Prevent repeated operations
- Ensure consistency

---

# Adapter Pattern

Notification Service uses abstraction:

```go
type EmailSender interface
```

Current implementation:

- SimulatedEmailSender

This allows replacing providers without changing business logic.

Possible future providers:

- SendGrid
- Mailgun
- AWS SES

Benefits:

- Loose coupling
- Extensibility
- Clean architecture

---

# Dockerized Environment

The project uses Docker Compose.

Containers:

- orders-db
- payments-db
- rabbitmq
- redis
- order-service
- payment-service
- notification-service

Benefits:

- Easy deployment
- Environment consistency
- Simplified setup

---

# How to Run the Project

## 1. Start containers

```bash
docker compose up --build
```

## 2. Check running containers

```bash
docker ps
```

## 3. Open RabbitMQ UI

```
http://localhost:15672
```

Credentials:

```
guest / guest
```

---

# API Testing

## Health Check

```
GET: http://localhost:8080/health
```

## Create Order

```
POST: http://localhost:8080/orders
```

Headers:

```
Content-Type: application/json
Idempotency-Key: test-1
```

Body:

```json
{
  "customer_id": "user-1",
  "item_name": "Laptop",
  "amount": 2000
}
```

## Get Order

```
GET: http://localhost:8080/orders/{id}
```

---

# Redis CLI Commands

Open Redis CLI:

```bash
docker exec -it redis redis-cli
```

Show all keys:

```
KEYS *
```

Check TTL:

```
TTL order:{id}
```

Read cached object:

```
GET order:{id}
```

---

# Demonstration Scenarios

## Cache-Aside

Expected logs:

```
[order-cache] MISS
[order-cache] SET
[order-cache] HIT
```

## Retry

Expected logs:

```
provider failed attempt=1/3
retry_in=2s
```

## Idempotency

Expected logs:

```
duplicate event ignored
```

---

# Architecture Concepts Used

- Microservices Architecture
- Event-Driven Architecture
- Cache-Aside Pattern
- Background Workers
- Adapter Pattern
- Retry with Exponential Backoff
- Idempotency
- Loose Coupling

---

# Conclusion

In Assignment 4, the system was improved to support production-oriented architecture patterns.

The project now supports:

- asynchronous processing
- distributed communication
- caching
- retry handling
- idempotency
- external provider abstraction

These improvements make the system:

- scalable
- reliable
- maintainable
- fault tolerant
- more production-ready
