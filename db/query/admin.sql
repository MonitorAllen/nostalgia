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