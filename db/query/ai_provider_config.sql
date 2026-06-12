-- name: GetAIProviderConfig :one
SELECT *
FROM ai_provider_configs
WHERE purpose = $1
LIMIT 1;

-- name: UpsertAIProviderConfig :one
INSERT INTO ai_provider_configs (
    purpose,
    provider,
    base_url,
    model,
    api_key_ciphertext,
    timeout_ms,
    max_input_chars,
    max_context_chars,
    max_suggestions,
    enabled,
    updated_by
) VALUES (
    sqlc.arg(purpose),
    sqlc.arg(provider),
    sqlc.arg(base_url),
    sqlc.arg(model),
    sqlc.arg(api_key_ciphertext),
    sqlc.arg(timeout_ms),
    sqlc.arg(max_input_chars),
    sqlc.arg(max_context_chars),
    sqlc.arg(max_suggestions),
    sqlc.arg(enabled),
    sqlc.narg(updated_by)
)
ON CONFLICT (purpose) DO UPDATE
SET provider = EXCLUDED.provider,
    base_url = EXCLUDED.base_url,
    model = EXCLUDED.model,
    api_key_ciphertext = EXCLUDED.api_key_ciphertext,
    timeout_ms = EXCLUDED.timeout_ms,
    max_input_chars = EXCLUDED.max_input_chars,
    max_context_chars = EXCLUDED.max_context_chars,
    max_suggestions = EXCLUDED.max_suggestions,
    enabled = EXCLUDED.enabled,
    updated_by = EXCLUDED.updated_by,
    updated_at = now()
RETURNING *;
