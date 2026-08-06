CREATE TABLE IF NOT EXISTS order_items (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id     UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id   UUID NOT NULL REFERENCES products(id),
    product_name TEXT NOT NULL,
    price        NUMERIC(12, 2) NOT NULL CHECK (price > 0),
    quantity     INT NOT NULL CHECK (quantity > 0),
    subtotal     NUMERIC(12, 2) NOT NULL CHECK (subtotal > 0),
    created_at   TIMESTAMPTZ DEFAULT now()
);
