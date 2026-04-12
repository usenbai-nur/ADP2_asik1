# AP2 Assignment 2 — gRPC Migration & Contract-First Development

## Student
- **Name:** Nurdaulet Usenbay
- **Group:** SE-2406

## Overview
This project is a migration of Assignment 1 from REST-based internal communication to **gRPC-based** internal communication while preserving **Clean Architecture**.

### Main idea
- **External API:** REST (`Order Service`)
- **Internal service-to-service communication:** gRPC (`Order Service -> Payment Service`)
- **Streaming:** `Order Service` exposes a server-side streaming gRPC method for real-time order status updates

## Repositories
### Main project repository
- `ADP2_asik1`

### Proto repository
- `ADP2_asik2_protos`

### Generated code repository
- `ADP2_asik2_generated`

## Architecture
### Services
#### Order Service
Responsibilities:
- Create order
- Get order by ID
- Cancel order
- Expose REST API for users
- Call Payment Service via gRPC
- Expose gRPC streaming for order status updates

#### Payment Service
Responsibilities:
- Process payment
- Store payment records
- Expose gRPC server

### Databases
- `orders_db` — used only by `Order Service`
- `payments_db` — used only by `Payment Service`

### Communication
- User -> REST -> Order Service
- Order Service -> gRPC -> Payment Service
- Stream Client -> gRPC stream -> Order Service

## Contract-First Flow
The project follows a contract-first approach:

1. `.proto` contracts are stored in `ADP2_asik2_protos`
2. GitHub Actions in `ADP2_asik2_generated` generates `.pb.go` files
3. Services import generated code from `ADP2_asik2_generated`

## Order Service REST API
### Create Order
`POST /orders`

Example request:
```json
{
  "customer_id": "cust-1",
  "item_name": "Keyboard",
  "amount": 15000
}
