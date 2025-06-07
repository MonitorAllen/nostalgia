-- name: CreateArticle :one
INSERT INTO articles (id,
                  title,
                  summary,
                  content,
                  is_publish,
                  owner)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetArticle :one
SELECT *
FROM articles
WHERE id = $1
LIMIT 1;

-- name: GetArticleForUpdate :one
SELECT * FROM articles
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;

-- name: ListArticles :many
SELECT p.id, p.title, p.summary, p.content, p.views, p.likes, p.is_publish, p.owner, p.created_at, p.updated_at, p.deleted_at, u.username,
       COALESCE(ARRAY_AGG(COALESCE(t.name, '')), ARRAY[]::TEXT[])::TEXT[] AS tags
FROM articles p
LEFT JOIN tags t on p.id = t.article_id
LEFT JOIN users u on p.owner = u.id
WHERE
    p.is_publish = sqlc.narg(is_publish)
GROUP BY p.id, p.title, p.summary, p.content, p.views, p.likes, p.is_publish, p.owner, p.created_at, p.updated_at, p.deleted_at, u.username
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountArticles :one
SELECT count(*)
FROM articles
where is_publish = sqlc.narg(is_publish);

-- name: UpdateArticle :one
UPDATE articles
SET
    title = COALESCE(sqlc.narg(title), title),
    summary = COALESCE(sqlc.narg(summary), summary),
    content = COALESCE(sqlc.narg(content), content),
    is_publish = COALESCE(sqlc.narg(is_publish), is_publish),
    updated_at = COALESCE(sqlc.narg(updated_at), updated_at)
WHERE id = sqlc.arg(id)
RETURNING *;;

-- name: ListAllArticles :many
SELECT id, title, summary, views, likes, is_publish, owner, created_at, updated_at, deleted_at
FROM articles
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAllArticles :one
SELECT count(*)
FROM articles;

-- name: DeleteArticle :exec
DELETE FROM articles WHERE id = $1;