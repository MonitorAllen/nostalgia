-- name: CreateComment :one
INSERT INTO comments
    (content, article_id, parent_id, from_user_id, to_user_id)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListCommentsByArticleID :many
SELECT c.*, from_u.username as from_user_name, to_u.username as to_user_name FROM comments c
LEFT JOIN users from_u on c.from_user_id = from_u.id
LEFT JOIN users to_u on c.to_user_id = to_u.id
WHERE
    c.article_id = $1 AND c.deleted_at = '0001-01-01 00:00:00.000000 +00:00'
ORDER BY c.id;

-- name: GetComment :one
SELECT * FROM comments
WHERE id = $1 LIMIT 1;

-- name: AddCommentLikes :one
UPDATE comments
SET likes = likes + 1
WHERE id = $1
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comments WHERE id = $1;

-- name: DeleteChildComments :exec
DELETE FROM comments WHERE parent_id = $1;