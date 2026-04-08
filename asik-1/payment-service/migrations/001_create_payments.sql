CREATE TABLE IF NOT EXISTS payments (
    id             TEXT        PRIMARY KEY,
    order_id       TEXT        NOT NULL,
    transaction_id TEXT        NOT NULL UNIQUE,
    amount         BIGINT      NOT NULL CHECK (amount > 0),   -- stored in cents, never float
    status         TEXT        NOT NULL,                       -- Authorized | Declined
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_status   ON payments(status);
