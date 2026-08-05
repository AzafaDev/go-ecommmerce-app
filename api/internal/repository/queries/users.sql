-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;
-- name: CreateUser :one
INSERT INTO users (full_name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;
-- name: GetUserByID :one
SELECT * 
FROM users
WHERE id = $1;
-- name: SetUserVerified :one
UPDATE users
SET email_verified_at = now(),
    updated_at = now()
WHERE id = $1
RETURNING *;
-- name: UpdatePasswordUser :one
UPDATE users
SET password_hash = $1,
    updated_at = now()
WHERE id = $2
RETURNING *;
-- name: UpdateUserRole :one
UPDATE users
SET role = $1,
    updated_at = now()
WHERE id = $2
RETURNING *;