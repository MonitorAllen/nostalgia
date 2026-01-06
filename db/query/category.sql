-- name: CreateCategory :one
INSERT INTO categories (name) VALUES ($1) RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1 AND is_system = false;

-- name: UpdateCategory :one
UPDATE categories
set
    name = $1,
    updated_at = now()
WHERE
    id = $2 RETURNING *;

-- name: ListAllCategories :many
SELECT * FROM categories ORDER BY created_at DESC;

-- name: ListCategoriesCountArticles :many
SELECT c.id, c.name, c.is_system, count(a.id) AS article_count, c.created_at, c.updated_at
FROM categories c
    LEFT JOIN articles a on a.category_id = c.id
    AND a.is_publish = true
GROUP BY
    c.id
ORDER BY article_count DESC, c.created_at DESC;

-- name: CountCategories :one
SELECT count(*) FROM categories;

-- name: ListArticlesByCategoryID :many
SELECT a.id, a.title, a.summary, a.views, a.likes, a.is_publish, a.owner, a.category_id, a.created_at, a.updated_at, a.deleted_at, c.name, u.username
FROM
    articles a
    LEFT JOIN categories c on a.category_id = c.id
    LEFT JOIN users u on a.owner = u.id
WHERE
    category_id = $1
    AND is_publish = true
ORDER BY a.created_at DESC
LIMIT $2
OFFSET
    $3;

-- name: CountArticlesByCategoryID :one
SELECT count(*) FROM articles WHERE category_id = $1;

-- name: GetCategory :one
SELECT * FROM categories WHERE id = @id LIMIT 1;

-- name: GetCategoryByName :one
SELECT * FROM categories WHERE name = $1 LIMIT 1;