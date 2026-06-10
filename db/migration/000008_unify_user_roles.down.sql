DROP INDEX IF EXISTS users_role_idx;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
