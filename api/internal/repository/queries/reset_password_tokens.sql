-- name: CreateResetPasswordToken :one
INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;
-- name: GetResetPasswordTokenByTokenHash :one
SELECT * FROM password_reset_tokens
WHERE token_hash = $1 AND expires_at > now();
-- name: DeletePasswordTokenByTokenHash :exec
DELETE FROM password_reset_tokens
WHERE token_hash = $1;