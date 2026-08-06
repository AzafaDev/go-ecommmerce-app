-- name: CreateProduct :one
INSERT INTO products (
        name,
        description,
        price,
        stock,
        sku,
        category,
        image_url
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
-- name: GetProductByID :one
SELECT *
FROM products
WHERE id = $1
    AND is_active = true;
-- name: AdminGetProductBySKU :one
SELECT *
FROM products
WHERE sku = $1;
-- name: AdminGetProductByID :one
SELECT *
FROM products
WHERE id = $1;
-- name: ListProducts :many
-- total_count rides along via window function, avoiding a separate COUNT query.
SELECT *, COUNT(*) OVER() AS total_count
FROM products
WHERE is_active = true
    AND (
        sqlc.narg('search')::text IS NULL
        OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    )
    AND (
        sqlc.narg('category')::text IS NULL
        OR category = sqlc.narg('category')::text
    )
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
-- name: AdminListProducts :many
SELECT *, COUNT(*) OVER() AS total_count
FROM products
WHERE (
        sqlc.narg('search')::text IS NULL
        OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    )
    AND (
        sqlc.narg('category')::text IS NULL
        OR category = sqlc.narg('category')::text
    )
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
-- name: CountProducts :one
-- Fallback for ListProducts when a page past the end returns no total_count row.
SELECT COUNT(*)
FROM products
WHERE is_active = true
    AND (
        sqlc.narg('search')::text IS NULL
        OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    )
    AND (
        sqlc.narg('category')::text IS NULL
        OR category = sqlc.narg('category')::text
    );
-- name: AdminCountProducts :one
-- Fallback for AdminListProducts; see CountProducts.
SELECT COUNT(*)
FROM products
WHERE (
        sqlc.narg('search')::text IS NULL
        OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    )
    AND (
        sqlc.narg('category')::text IS NULL
        OR category = sqlc.narg('category')::text
    );
-- name: UpdateProduct :one
UPDATE products
SET name = $1,
    description = $2,
    price = $3,
    stock = $4,
    category = $5,
    is_active = $6,
    image_url = $7,
    updated_at = now()
WHERE id = $8
RETURNING *;
-- name: SoftDeleteProduct :one
UPDATE products
SET is_active = false
WHERE id = $1
RETURNING *;
-- name: DecrementProductStock :one
UPDATE products
SET stock = stock - $1,
    updated_at = now()
WHERE id = $2
    AND stock >= $1
RETURNING *;
-- name: IncrementProductStock :exec
UPDATE products
SET stock = stock + $1,
    updated_at = now()
WHERE id = $2;
-- name: GetDistinctCategories :many
SELECT DISTINCT category
FROM products
WHERE is_active = true
ORDER BY category;
-- name: AdminGetDistinctCategories :many
SELECT DISTINCT category
FROM products
ORDER BY category;
