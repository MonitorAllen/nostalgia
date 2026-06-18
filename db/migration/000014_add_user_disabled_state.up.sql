ALTER TABLE users
ADD COLUMN disabled_at timestamptz,
ADD COLUMN disabled_reason text NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS users_role_disabled_created_idx
ON users(role, disabled_at, created_at DESC);
