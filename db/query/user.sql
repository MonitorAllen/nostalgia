-- name: CreateUser :one
INSERT INTO users (
    id,
    username,
    hashed_password,
    full_name,
    email
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: CountAdminUsers :one
SELECT count(*) FROM users WHERE role = 'admin';

-- name: GetFirstAdminUser :one
SELECT *
FROM users
WHERE role = 'admin'
ORDER BY created_at ASC
LIMIT 1;

-- name: CreateUserWithRole :one
INSERT INTO users (
    id,
    username,
    hashed_password,
    full_name,
    email,
    is_email_verified,
    role
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET
    hashed_password = COALESCE(sqlc.narg(hashed_password), hashed_password),
    full_name = COALESCE(sqlc.narg(full_name), full_name),
    email = COALESCE(sqlc.narg(email), email),
    is_email_verified=COALESCE(sqlc.narg(is_email_verified), is_email_verified)
WHERE
    id = sqlc.arg(id)
RETURNING *;

-- name: ListAdminUsers :many
SELECT
    id,
    username,
    full_name,
    email,
    is_email_verified,
    role,
    created_at,
    updated_at,
    disabled_at,
    disabled_reason
FROM users
WHERE role = 'visitor'
  AND (
    sqlc.arg(status)::text = 'all'
    OR (sqlc.arg(status)::text = 'enabled' AND disabled_at IS NULL)
    OR (sqlc.arg(status)::text = 'disabled' AND disabled_at IS NOT NULL)
  )
  AND (
    sqlc.narg(q)::text IS NULL
    OR username ILIKE '%' || sqlc.narg(q)::text || '%'
    OR full_name ILIKE '%' || sqlc.narg(q)::text || '%'
    OR email ILIKE '%' || sqlc.narg(q)::text || '%'
  )
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAdminUsersByFilter :one
SELECT count(*)
FROM users
WHERE role = 'visitor'
  AND (
    sqlc.arg(status)::text = 'all'
    OR (sqlc.arg(status)::text = 'enabled' AND disabled_at IS NULL)
    OR (sqlc.arg(status)::text = 'disabled' AND disabled_at IS NOT NULL)
  )
  AND (
    sqlc.narg(q)::text IS NULL
    OR username ILIKE '%' || sqlc.narg(q)::text || '%'
    OR full_name ILIKE '%' || sqlc.narg(q)::text || '%'
    OR email ILIKE '%' || sqlc.narg(q)::text || '%'
  );

-- name: UpdateVisitorUser :one
UPDATE users
SET
    full_name = sqlc.arg(full_name),
    email = sqlc.arg(email),
    is_email_verified = sqlc.arg(is_email_verified),
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND role = 'visitor'
RETURNING *;

-- name: DisableVisitorUser :one
UPDATE users
SET
    disabled_at = COALESCE(disabled_at, now()),
    disabled_reason = sqlc.arg(disabled_reason),
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND role = 'visitor'
RETURNING *;

-- name: EnableVisitorUser :one
UPDATE users
SET
    disabled_at = NULL,
    disabled_reason = '',
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND role = 'visitor'
RETURNING *;
