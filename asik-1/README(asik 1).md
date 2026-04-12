# AP2 Assignment 1 — Clean Architecture Microservices (Order & Payment)

## Architecture Overview

This project implements two independent microservices in Go, following **Clean Architecture** principles:

- **Order Service** (`:8080`) — manages orders, owns the `orders_db` database
- **Payment Service** (`:8081`) — processes payments, owns the `payments_db` database

Services communicate only via REST. Each service has its own database (**Database per Service** pattern).

### Bounded Contexts
- **Order** knows only about orders and their statuses (`Pending → Paid / Failed / Cancelled`)
- **Payment** knows only about payment transactions. The only link is `order_id` as a reference (no JOINs)

### Layers (Clean Architecture)
```
domain/      ← domain models and interfaces (Ports), no external dependencies
usecase/     ← business logic, depends only on domain
repository/  ← port implementations (PostgreSQL, HTTP client)
transport/   ← HTTP handlers (Gin), thin layer without logic
app/         ← configuration and DB initialization
cmd/         ← entry point (Composition Root, manual DI)
```

### Failure Handling
If Payment Service is unavailable:
- 2 second timeout — service does not hang indefinitely
- Order is marked as `Failed`
- Order Service returns `503 Service Unavailable`

---

## Requirements

- Go 1.22+
- PostgreSQL 14+

---

## Running the Services

Open **two terminals**:

**Terminal 1 — Payment Service:**
```
cd payment-service
go run ./cmd/payment-service/main.go
```

**Terminal 2 — Order Service:**
```
cd order-service
go run ./cmd/order-service/main.go
```

---

## API — Request Examples (Postman)

### Create Order
- **Method:** POST
- **URL:** `http://localhost:8080/orders`
- **Body (JSON):**
  ```json
  {
    "customer_id": "user-1",
    "item_name": "telefon",
    "amount": 5000
  }
  ```

### Get Order by ID
- **Method:** GET
- **URL:** `http://localhost:8080/orders/{id}`

### Cancel Order (only Pending)
- **Method:** PATCH
- **URL:** `http://localhost:8080/orders/{id}/cancel`

### Get Payment by order_id
- **Method:** GET
- **URL:** `http://localhost:8081/payments/{order_id}`

### Test Limit — amount > 100000 → Declined
- **Method:** POST
- **URL:** `http://localhost:8080/orders`
- **Body (JSON):**
  ```json
  {
    "customer_id": "user-2",
    "item_name": "ukulele",
    "amount": 200000
  }
  ```

### Test Idempotency (Bonus)
- **Method:** POST
- **URL:** `http://localhost:8080/orders`
- **Headers:**
  - `Idempotency-Key: my-unique-key-001`
- **Body (JSON):**
  ```json
  {
    "customer_id": "user-3",
    "item_name": "pencil",
    "amount": 30
  }
  ```

---

## Business Rules

| Rule | Description |
|---|---|
| `amount > 0` | Amount must be positive |
| `amount` in cents | `1000` = $10.00, no `float64` |
| `amount > 100000` | Payment Service declines (`Declined`) |
| Cancel | Only `Pending` orders can be cancelled |
| Timeout | Order Service waits for Payment Service max 2 seconds |
| Unavailable | If Payment unavailable → order `Failed`, response `503` |
| Idempotency | `Idempotency-Key` header prevents duplicates |
