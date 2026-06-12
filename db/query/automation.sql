-- name: CreateAutomationArticleRequest :one
INSERT INTO automation_article_requests (
  idempotency_key,
  request_hash,
  key_id,
  status,
  title,
  source_topic,
  source_prompt,
  generation_model,
  error_message,
  client_ip,
  user_agent
) VALUES (
  sqlc.arg(idempotency_key),
  sqlc.arg(request_hash),
  sqlc.arg(key_id),
  sqlc.arg(status),
  sqlc.arg(title),
  sqlc.arg(source_topic),
  sqlc.arg(source_prompt),
  sqlc.arg(generation_model),
  sqlc.arg(error_message),
  sqlc.arg(client_ip),
  sqlc.arg(user_agent)
)
RETURNING *;

-- name: GetAutomationArticleRequestByIdempotencyKey :one
SELECT *
FROM automation_article_requests
WHERE idempotency_key = $1
LIMIT 1;

-- name: MarkAutomationArticleRequestCreated :one
UPDATE automation_article_requests
SET status = 'created',
    article_id = sqlc.arg(article_id),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CountAutomationDraftsToday :one
SELECT count(*)
FROM automation_article_requests
WHERE status = 'created'
  AND created_at >= date_trunc('day', now());
