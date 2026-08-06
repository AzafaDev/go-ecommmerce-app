-- name: AddCartItem :one
-- Zero rows = product not found/inactive; quantity = NULL = insufficient stock.
WITH target_product AS (
    SELECT products.id, products.name, products.price, products.image_url, products.stock
    FROM products
    WHERE products.id = sqlc.arg('product_id') AND products.is_active = true
),
upsert AS (
    INSERT INTO cart_items (user_id, product_id, quantity)
    SELECT sqlc.arg('user_id'), sqlc.arg('product_id'), sqlc.arg('quantity')
    WHERE sqlc.arg('quantity')::int <= (SELECT stock FROM target_product)
    ON CONFLICT (user_id, product_id)
    DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity,
                  updated_at = now()
    WHERE cart_items.quantity + EXCLUDED.quantity <= (SELECT stock FROM target_product)
    RETURNING quantity
)
SELECT
    tp.id AS product_id,
    tp.name,
    tp.price,
    tp.image_url,
    u.quantity
FROM target_product tp
LEFT JOIN upsert u ON true;
-- name: UpdateCartItemQuantity :one
-- Zero rows = product not found/inactive; item_exists = false = no cart row yet; quantity = NULL = insufficient stock.
WITH target_product AS (
    SELECT products.id, products.name, products.price, products.image_url, products.stock
    FROM products
    WHERE products.id = sqlc.arg('product_id') AND products.is_active = true
),
existing_item AS (
    SELECT 1 AS found FROM cart_items
    WHERE cart_items.user_id = sqlc.arg('user_id') AND cart_items.product_id = sqlc.arg('product_id')
),
updated AS (
    UPDATE cart_items
    SET quantity = sqlc.arg('quantity'),
        updated_at = now()
    WHERE user_id = sqlc.arg('user_id')
        AND product_id = sqlc.arg('product_id')
        AND sqlc.arg('quantity')::int <= (SELECT stock FROM target_product)
    RETURNING quantity
)
SELECT
    tp.id AS product_id,
    tp.name,
    tp.price,
    tp.image_url,
    EXISTS (SELECT 1 FROM existing_item) AS item_exists,
    u.quantity
FROM target_product tp
LEFT JOIN updated u ON true;
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
