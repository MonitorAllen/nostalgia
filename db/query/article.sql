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

-- name: ListArticles :many
SELECT p.id, p.title, p.summary, p.content, p.views, p.likes, p.is_publish, p.owner, p.create_at, p.update_at, p.delete_at, u.username,
       COALESCE(ARRAY_AGG(t.name), '{}') AS tags
FROM articles p
LEFT JOIN tags t on p.id = t.article_id
LEFT JOIN users u on p.owner = u.id
WHERE
    p.is_publish = sqlc.narg(is_publish)
GROUP BY p.id, p.title, p.summary, p.content, p.views, p.likes, p.is_publish, p.owner, p.create_at, p.update_at, p.delete_at, u.username
ORDER BY p.create_at DESC
LIMIT $1 OFFSET $2;

-- name: CountArticles :one
SELECT count(*)
FROM articles
where is_publish = $1;

-- name: UpdateArticle :one
UPDATE articles
SET
    title = COALESCE(sqlc.narg(title), title),
    summary = COALESCE(sqlc.narg(summary), summary),
    content = COALESCE(sqlc.narg(content), content),
    is_publish = COALESCE(sqlc.narg(is_publish), is_publish),
    update_at = COALESCE(sqlc.narg(update_at), update_at)
WHERE id = sqlc.arg(id)
RETURNING *;;
