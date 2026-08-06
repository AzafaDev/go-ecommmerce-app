CREATE TABLE IF NOT EXISTS cart_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity    INT NOT NULL CHECK (quantity > 0),
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now(),
    UNIQUE (user_id, product_id)
);
