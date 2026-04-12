CREATE TABLE IF NOT EXISTS payments (
    id              TEXT        PRIMARY KEY,
    order_id        TEXT        NOT NULL,
    transaction_id  TEXT        NOT NULL UNIQUE,
    amount          BIGINT      NOT NULL CHECK (amount > 0),
    status          TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);