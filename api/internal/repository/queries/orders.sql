-- name: CreateOrder :one
INSERT INTO orders (user_id, status, total_amount)
VALUES ($1, $2, $3)
RETURNING *;
-- name: GetOrderByID :one
SELECT *
FROM orders
WHERE id = $1;
-- name: GetOrderByMidtransID :one
SELECT *
FROM orders
WHERE midtrans_order_id = $1;
-- name: ListOrdersByUser :many
SELECT *
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC;
-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $1,
    paid_at = $2,
    updated_at = now()
WHERE id = $3
RETURNING *;
-- name: SetOrderPaymentInfo :one
UPDATE orders
SET midtrans_order_id = $1,
    snap_token = $2,
    updated_at = now()
WHERE id = $3
RETURNING *;
-- name: ListExpiredPendingOrders :many
SELECT *
FROM orders
WHERE status = 'pending_payment'
    AND created_at < $1;
