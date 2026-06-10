UPDATE users SET role = 'visitor' WHERE role NOT IN ('admin', 'visitor');

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'visitor'));

CREATE INDEX IF NOT EXISTS users_role_idx ON users(role);
