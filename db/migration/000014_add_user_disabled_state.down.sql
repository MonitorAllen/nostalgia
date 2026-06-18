DROP INDEX IF EXISTS users_role_disabled_created_idx;

ALTER TABLE users
DROP COLUMN IF EXISTS disabled_reason,
DROP COLUMN IF EXISTS disabled_at;
