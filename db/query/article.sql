-- name: CreateArticle :one
INSERT INTO articles (id,
                  title,
                  summary,
                  content,
                  is_publish,
                  owner,
                  category_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetArticle :one
SELECT a.*, c.name as category_name
FROM articles a
LEFT OUTER JOIN categories c on c.id = a.category_id
WHERE a.id = $1
LIMIT 1;

-- name: GetArticleForUpdate :one
SELECT * FROM articles
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;

-- name: ListArticles :many
SELECT a.id, a.title, a.summary, a.views, a.likes, a.is_publish, a.owner, a.created_at, a.updated_at, a.deleted_at, c.name as category_name, u.username
FROM articles a
LEFT JOIN categories c on c.id = a.category_id
LEFT JOIN users u on a.owner = u.id
WHERE
    a.is_publish = COALESCE(sqlc.narg(is_publish), a.is_publish)
  AND a.category_id = COALESCE(sqlc.narg(category_id), a.category_id)
ORDER BY a.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountArticles :one
SELECT count(*)
FROM articles
WHERE is_publish = COALESCE(sqlc.narg(is_publish), is_publish)
  AND category_id = COALESCE(sqlc.narg(category_id), category_id);

-- name: UpdateArticle :one
UPDATE articles
SET
    title = COALESCE(sqlc.narg(title), title),
    summary = COALESCE(sqlc.narg(summary), summary),
    content = COALESCE(sqlc.narg(content), content),
    is_publish = COALESCE(sqlc.narg(is_publish), is_publish),
    category_id = COALESCE(sqlc.narg(category_id), category_id),
    updated_at = COALESCE(sqlc.narg(updated_at), updated_at)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: IncrementArticleLikes :exec
UPDATE articles SET likes = likes + 1 WHERE id = @id;

-- name: IncrementArticleViews :exec
UPDATE articles SET views = views + 1 WHERE  id = @id;

-- name: ListAllArticles :many
SELECT a.id, title, summary, views, likes, is_publish, owner, a.created_at, a.updated_at, deleted_at, c.name as category_name
FROM articles a
LEFT JOIN categories c on c.id = a.category_id
ORDER BY a.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAllArticles :one
SELECT count(*)
FROM articles;

-- name: DeleteArticle :exec
DELETE FROM articles WHERE id = $1;

-- name: SetArticleDefaultCategoryIdByCategoryId :exec
UPDATE articles SET category_id = 1 WHERE category_id = $1;