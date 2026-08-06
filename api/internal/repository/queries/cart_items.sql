-- name: AddCartItem :one
INSERT INTO cart_items (user_id, product_id, quantity)
SELECT $1, $2, $3
WHERE $3 <= (SELECT stock FROM products WHERE id = $2)
ON CONFLICT (user_id, product_id)
DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity,
              updated_at = now()
WHERE cart_items.quantity + EXCLUDED.quantity <= (SELECT stock FROM products WHERE id = EXCLUDED.product_id)
RETURNING *;
-- name: UpdateCartItemQuantity :one
UPDATE cart_items
SET quantity = $1,
    updated_at = now()
WHERE user_id = $2
    AND product_id = $3
RETURNING *;
-- name: DeleteCartItem :execrows
DELETE FROM cart_items
WHERE user_id = $1
    AND product_id = $2;
-- name: ClearCart :exec
DELETE FROM cart_items
WHERE user_id = $1;
-- name: ListCartItems :many
SELECT
    ci.product_id,
    ci.quantity,
    p.name,
    p.price,
    p.image_url,
    p.is_active,
    p.stock,
    (ci.quantity * p.price)::numeric AS subtotal
FROM cart_items ci
JOIN products p ON p.id = ci.product_id
WHERE ci.user_id = $1
ORDER BY ci.created_at DESC;
