# Architecture Diagram

```mermaid
flowchart LR
    U[User / Postman] -->|REST| O[Order Service]
    O -->|SQL| ODB[(orders_db)]

    O -->|gRPC ProcessPayment| P[Payment Service]
    P -->|SQL| PDB[(payments_db)]

    C[Streaming Client] -->|gRPC SubscribeToOrderUpdates| O

    PR[ADP2_asik2_protos] -->|Proto Contracts| GR[ADP2_asik2_generated]
    GR -->|Generated Go Code| O
    GR -->|Generated Go Code| P