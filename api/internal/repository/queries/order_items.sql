-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, product_name, price, quantity, subtotal)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
-- name: ListOrderItemsByOrder :many
SELECT *
FROM order_items
WHERE order_id = $1
ORDER BY created_at ASC;
