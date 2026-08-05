CREATE TABLE IF NOT EXISTS email_verification_tokens(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT fk_email_verification_tokens_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE index idx_email_verifications_tokens_user_id ON email_verification_tokens(user_id);
ALTER TABLE users
ADD COLUMN email_verified_at TIMESTAMPTZ;