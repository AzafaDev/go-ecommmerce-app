-- name: CreateProduct :one
INSERT INTO products (
        name,
        description,
        price,
        stock,
        sku,
        category
    )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
-- name: GetProductByID :one
SELECT *
FROM products
WHERE id = $1
    AND is_active = true;
-- name: AdminGetProductByID :one
SELECT *
FROM products
WHERE id = $1;
-- name: ListProducts :many
SELECT *
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
SELECT *
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
    updated_at = now()
WHERE id = $7
RETURNING *;
-- name: SoftDeleteProduct :one
UPDATE products
SET is_active = false
WHERE id = $1
RETURNING *;
-- name: GetDistinctCategories :many
SELECT DISTINCT category
FROM products
WHERE is_active = true
ORDER BY category;
-- name: AdminGetDistinctCategories :many
SELECT DISTINCT category
FROM products
ORDER BY category;
