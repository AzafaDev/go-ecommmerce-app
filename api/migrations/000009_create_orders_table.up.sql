CREATE TABLE IF NOT EXISTS orders (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status           TEXT NOT NULL DEFAULT 'pending_payment'
                     CHECK (status IN ('pending_payment', 'paid', 'cancelled', 'expired')),
    total_amount     NUMERIC(12, 2) NOT NULL CHECK (total_amount > 0),
    midtrans_order_id TEXT,
    snap_token       TEXT,
    created_at       TIMESTAMPTZ DEFAULT now(),
    updated_at       TIMESTAMPTZ DEFAULT now(),
    paid_at          TIMESTAMPTZ
);
