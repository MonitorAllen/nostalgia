-- name: CreateAdmin :one
INSERT INTO admins (
    username,
    hashed_password,
    is_active,
    role_id
) VALUES (
             $1, $2, $3, $4
         ) RETURNING *;

-- name: GetAdmin :one
SELECT * FROM admins
WHERE username = $1 LIMIT 1;

-- name: UpdateAdmin :one
UPDATE admins
SET
    username = COALESCE(sqlc.narg(username), username),
    hashed_password = COALESCE(sqlc.narg(hashed_password), hashed_password),
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    role_id = COALESCE(sqlc.narg(role_id), role_id),
    updated_at = COALESCE(sqlc.narg(updated_at), updated_at)
WHERE id = sqlc.arg(id)
RETURNING *;;

-- name: GetAdminById :one
SELECT * FROM admins
WHERE id = $1 LIMIT 1;