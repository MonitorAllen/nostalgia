-- name: CreateArticle :one
INSERT INTO articles (id,
                      title,
                      summary,
                      content,
                      is_publish,
                      owner,
                      category_id,
                      cover)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetArticle :one
SELECT a.id,
       a.title,
       a.summary,
       a.content,
       a.is_publish,
       a.views,
       a.likes,
       a.cover,
       a.slug,
       a.check_outdated,
       a.last_updated,
       a.read_time,
       a.owner,
       a.created_at,
       a.updated_at,
       a.deleted_at,
       a.category_id,
       c.name as category_name
FROM articles a
         LEFT JOIN categories c on c.id = a.category_id
WHERE a.id = $1
LIMIT 1;

-- name: GetArticleBySlug :one
SELECT a.id,
       a.title,
       a.summary,
       a.content,
       a.is_publish,
       a.views,
       a.likes,
       a.cover,
       a.slug,
       a.check_outdated,
       a.last_updated,
       a.read_time,
       a.owner,
       a.created_at,
       a.updated_at,
       a.deleted_at,
       a.category_id,
       c.name as category_name
FROM articles a
         LEFT JOIN categories c on c.id = a.category_id
WHERE a.slug = $1
LIMIT 1;

-- name: GetArticleForUpdate :one
SELECT id,
       title,
       summary,
       content,
       views,
       likes,
       is_publish,
       owner,
       created_at,
       updated_at,
       deleted_at,
       category_id
FROM articles
WHERE id = $1
LIMIT 1 FOR NO KEY UPDATE;

-- name: ListArticles :many
SELECT a.id,
       a.title,
       a.summary,
       a.views,
       a.likes,
       a.is_publish,
       a.cover,
       a.slug,
       a.check_outdated,
       a.last_updated,
       a.read_time,
       a.owner,
       a.created_at,
       a.updated_at,
       a.deleted_at,
       c.name as category_name,
       u.username
FROM articles a
         LEFT JOIN categories c on c.id = a.category_id
         LEFT JOIN users u on a.owner = u.id
WHERE a.is_publish = COALESCE(sqlc.narg(is_publish), a.is_publish)
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
SET title          = COALESCE(sqlc.narg(title), title),
    summary        = COALESCE(sqlc.narg(summary), summary),
    content        = COALESCE(sqlc.narg(content), content),
    is_publish     = COALESCE(sqlc.narg(is_publish), is_publish),
    category_id    = COALESCE(sqlc.narg(category_id), category_id),
    cover          = COALESCE(sqlc.narg(cover), cover),
    slug           = COALESCE(sqlc.narg(slug), slug),
    check_outdated = COALESCE(sqlc.narg(check_outdated), check_outdated),
    last_updated   = COALESCE(sqlc.narg(last_updated), last_updated),
    read_time   = COALESCE(sqlc.narg(read_time), read_time),
    updated_at     = COALESCE(sqlc.narg(updated_at), updated_at)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: IncrementArticleLikes :exec
UPDATE articles
SET likes = likes + 1
WHERE id = @id;

-- name: IncrementArticleViews :exec
UPDATE articles
SET views = views + 1
WHERE id = @id;

-- name: ListAllArticles :many
SELECT a.id,
       title,
       summary,
       views,
       likes,
       is_publish,
       a.cover,
       a.slug,
       a.check_outdated,
       a.last_updated,
       a.read_time,
       owner,
       a.created_at,
       a.updated_at,
       deleted_at,
       c.name as category_name
FROM articles a
         LEFT JOIN categories c on c.id = a.category_id
ORDER BY a.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAllArticles :one
SELECT count(*)
FROM articles;

-- name: DeleteArticle :exec
DELETE
FROM articles
WHERE id = $1;

-- name: SetArticleDefaultCategoryIdByCategoryId :exec
UPDATE articles
SET category_id = 1
WHERE category_id = $1;

-- name: SearchArticles :many
SELECT a.id,
       a.title,
       a.summary,
       a.views,
       a.likes,
       a.is_publish,
       a.cover,
       a.slug,
       a.check_outdated,
       a.last_updated,
       a.read_time,
       a.owner,
       a.created_at,
       a.updated_at,
       a.deleted_at,
       c.name                             as category_name,
       u.username,
       pgroonga_score(a.tableoid, a.ctid) AS score
FROM articles a
         LEFT JOIN categories c on c.id = a.category_id
         LEFT JOIN users u on a.owner = u.id
WHERE (title || ' ' || summary || ' ' || content) &@~ sqlc.arg(keyword)::text
  AND (sqlc.narg('is_publish')::boolean IS NULL OR a.is_publish = sqlc.narg('is_publish'))
  AND a.deleted_at = '0001-01-01 00:00:00Z'
ORDER BY score DESC, a.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountSearchArticles :one
SELECT count(*)
FROM articles a
WHERE (title || ' ' || summary || ' ' || content) &@~ sqlc.arg(keyword)::text
  AND (sqlc.narg('is_publish')::boolean IS NULL OR a.is_publish = sqlc.narg('is_publish'))
  AND a.deleted_at = '0001-01-01 00:00:00Z';