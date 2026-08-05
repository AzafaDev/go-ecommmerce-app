-- name: CreateVericationEmail :one
INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetEmailVerificationByTokenHash :one
SELECT * FROM email_verification_tokens
WHERE token_hash = $1 AND expires_at > now();

-- name: DeleteEmailVerificationByTokenHash :exec
DELETE FROM email_verification_tokens
WHERE token_hash = $1;