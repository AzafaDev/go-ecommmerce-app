-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3);
-- name: GetRefreshToken :one
SELECT *
FROM refresh_tokens
WHERE token_hash = $1
    AND expires_at > now()
    AND revoked_at IS NULL;
-- name: GetRefreshTokenAnyStatus :one
SELECT *
FROM refresh_tokens
WHERE token_hash = $1;
-- name: RevokeRefreshTokenByUserID :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE user_id = $1
    AND revoked_at IS NULL
    AND expires_at > now();
-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE token_hash = $1
    AND revoked_at IS NULL;