CREATE TABLE IF NOT EXISTS orders (
    order_uid   TEXT PRIMARY KEY,
    payload     JSONB NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT now()
);


CREATE INDEX IF NOT EXISTS idx_orders_order_uid ON orders (order_uid);