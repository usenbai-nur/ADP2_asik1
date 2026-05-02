#  Architecture Diagram 

```text
          ┌───────────────┐
          │    Client     │
          └──────┬────────┘
                 │ HTTP
                 ▼
          ┌───────────────┐
          │ Order Service │
          └──────┬────────┘
                 │ gRPC
                 ▼
          ┌───────────────┐
          │ Payment       │
          │ Service       │
          └──────┬────────┘
                 │ Publish Event
                 ▼
          ┌───────────────┐
          │  RabbitMQ     │
          └──────┬────────┘
                 │ Consume Event
                 ▼
          ┌───────────────┐
          │ Notification  │
          │ Service       │
          └───────────────┘
