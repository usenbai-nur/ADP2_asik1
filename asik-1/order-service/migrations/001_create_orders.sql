CREATE TABLE IF NOT EXISTS orders (
    id               TEXT        PRIMARY KEY,
    customer_id      TEXT        NOT NULL,
    item_name        TEXT        NOT NULL,
    amount           BIGINT      NOT NULL CHECK (amount > 0),   -- stored in cents
    status           TEXT        NOT NULL DEFAULT 'Pending',     -- Pending | Paid | Failed | Cancelled
    idempotency_key  TEXT        UNIQUE,                         -- nullable; UNIQUE prevents duplicate orders
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_status      ON orders(status);
