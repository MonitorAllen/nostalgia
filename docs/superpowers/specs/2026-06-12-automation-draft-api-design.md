# Automation Draft API Design

## Background

Nostalgia needs a controlled way for Codex automation to create blog article drafts. Codex will generate the complete draft package, including title, summary, content, category, slug, cover, and source metadata. Nostalgia will not ask Codex to modify the repository and will not automatically publish generated content.

The first version focuses on a secure write entrypoint, durable auditability, idempotent retries, email notification, and a clear backend review marker.

## Goals

- Expose a production-safe API for Codex automation to create article drafts.
- Require HMAC request signing, timestamp validation, and idempotency.
- Force all automation-created articles to remain unpublished until reviewed by the owner.
- Record every accepted automation request for audit and retry behavior.
- Notify the owner by email when a draft is created or when an authenticated request fails business validation.
- Mark automation-created drafts in the backend article list.
- Keep the design compatible with the existing Go/Gin/gRPC-Gateway backend, PostgreSQL/sqlc data layer, Asynq worker, mail module, and Vue backend UI.

## Non-Goals

- Do not let Codex automation publish articles directly.
- Do not let Codex automation create, update, or delete categories in this phase.
- Do not call OpenAI or Codex from the Nostalgia backend in this phase.
- Do not introduce a full review workflow with approvals, rejection comments, or revision history in this phase.
- Do not implement AI polishing, SEO/GEO generation, or scheduled topic planning in this phase.
- Do not expose automation secrets in the frontend or API responses.

## High-Level Flow

```text
Codex Automation
  -> POST /api/automation/articles/drafts
  -> HMAC authentication, timestamp check, body hash check, idempotency check
  -> validate category, slug, payload size, required fields, daily draft limit
  -> create unpublished article marked as automation draft
  -> record automation request audit row
  -> enqueue owner email notification
  -> return article draft ID and backend edit URL
```

The API creates a draft synchronously after authentication and validation. Email notification is asynchronous through the existing worker infrastructure.

## API Contract

### Endpoint

```http
POST /api/automation/articles/drafts
```

This endpoint lives under the public Gin API surface but uses dedicated automation authentication. It must not use user JWT authentication because Codex automation should not need an owner session token.

### Required Headers

```text
X-Automation-Key-Id: codex-daily-writer
X-Automation-Timestamp: 2026-06-12T10:30:00+08:00
X-Automation-Signature: v1=<hex-encoded-hmac-sha256>
Idempotency-Key: 2026-06-12-codex-daily-writer-go-cache
Content-Type: application/json
```

### Request Body

```json
{
  "title": "文章标题",
  "summary": "文章摘要",
  "content": "<h2>...</h2><p>...</p>",
  "category_id": 1,
  "slug": "optional-custom-slug",
  "cover": "/images/go.png",
  "check_outdated": true,
  "source_topic": "今天的创作主题",
  "source_prompt": "Codex 本次使用的创作提示词",
  "generation_model": "codex/gpt-5.x"
}
```

### Request Validation

- `title` is required, trimmed, and limited to 160 characters.
- `summary` is required, trimmed, and limited to 500 characters.
- `content` is required and must not exceed the configured request body limit.
- `category_id` is required and must refer to an existing category.
- `slug` is optional. When present, it must follow the existing public slug constraints and must be unique.
- `cover` is optional. When empty, the existing default cover is used.
- `check_outdated` defaults to `true` when omitted.
- `source_topic`, `source_prompt`, and `generation_model` are optional audit metadata and must be size-limited.
- The backend ignores any `is_publish`, `owner`, `created_by_automation`, or `automation_status` values if a client sends them.

### Response: Created

```json
{
  "article": {
    "id": "uuid",
    "title": "文章标题",
    "is_publish": false,
    "created_by_automation": true,
    "automation_status": "pending_review"
  },
  "review_url": "https://example.com/backend/articles/uuid",
  "idempotency_key": "2026-06-12-codex-daily-writer-go-cache",
  "status": "created"
}
```

### Response: Idempotent Replay

When the same `Idempotency-Key` and request body hash are submitted again after a successful creation, return `200 OK` with the same article ID and `status: "replayed"`.

### Response: Conflict

Return `409 Conflict` when:

- the same `Idempotency-Key` is used with a different request body hash;
- the requested slug already belongs to another article;
- the daily automation draft limit has been reached.

## HMAC Authentication

The signature base string is:

```text
METHOD + "\n" +
PATH + "\n" +
TIMESTAMP + "\n" +
IDEMPOTENCY_KEY + "\n" +
SHA256(BODY)
```

Example path value:

```text
/api/automation/articles/drafts
```

The backend calculates:

```text
hex(HMAC-SHA256(secret, signature_base_string))
```

The expected header is:

```text
X-Automation-Signature: v1=<hex-digest>
```

Validation rules:

- `X-Automation-Key-Id`, `X-Automation-Timestamp`, `X-Automation-Signature`, and `Idempotency-Key` are required.
- `X-Automation-Key-Id` must match the configured key ID.
- `X-Automation-Timestamp` must parse as RFC3339.
- Timestamp skew must be within `AUTOMATION_SIGNATURE_TTL`, default `5m`.
- Signature comparison must use constant-time comparison.
- The request body must be read exactly once, hashed, then restored for JSON binding.
- Authentication failures return `401 Unauthorized` with a generic message and do not enqueue email notifications.

## Configuration

Add configuration fields to `util.Config`:

```env
AUTOMATION_HMAC_KEY_ID=codex-daily-writer
AUTOMATION_HMAC_SECRET=replace-with-a-random-secret
AUTOMATION_SIGNATURE_TTL=5m
AUTOMATION_DAILY_DRAFT_LIMIT=1
AUTOMATION_NOTIFY_EMAIL=owner@example.com
```

Configuration behavior:

- Automation draft API is disabled when `AUTOMATION_HMAC_KEY_ID` or `AUTOMATION_HMAC_SECRET` is empty.
- `AUTOMATION_SIGNATURE_TTL` defaults to `5m`.
- `AUTOMATION_DAILY_DRAFT_LIMIT` defaults to `1`.
- `AUTOMATION_NOTIFY_EMAIL` defaults to `EMAIL_SENDER_ADDRESS` only for local development and tests; production deployments should set it explicitly.
- `.env.example`, README, Docker Compose, and CI test env should document names only and use placeholder secrets.

## Data Model

### Articles Table Additions

Add lightweight automation fields to `articles`:

```sql
created_by_automation boolean NOT NULL DEFAULT false;
automation_status varchar(32) NOT NULL DEFAULT '';
automation_request_id bigint;
```

Allowed first-version statuses:

- empty string for normal manually created articles;
- `pending_review` for automation-created drafts waiting for owner review;
- `published` after the owner publishes the article.

The article still uses existing `is_publish` as the public visibility source of truth. Automation-created articles must always start with `is_publish=false`.

### Automation Request Audit Table

Add `automation_article_requests`:

```sql
id bigserial PRIMARY KEY;
idempotency_key varchar(160) UNIQUE NOT NULL;
request_hash varchar(64) NOT NULL;
key_id varchar(100) NOT NULL;
status varchar(32) NOT NULL;
article_id uuid;
title varchar NOT NULL DEFAULT '';
source_topic varchar NOT NULL DEFAULT '';
source_prompt text NOT NULL DEFAULT '';
generation_model varchar NOT NULL DEFAULT '';
error_message text NOT NULL DEFAULT '';
client_ip varchar NOT NULL DEFAULT '';
user_agent varchar NOT NULL DEFAULT '';
created_at timestamptz NOT NULL DEFAULT now();
updated_at timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z';
```

`status` values:

- `created`: draft was created successfully;
- `failed_validation`: authentication passed, but business validation failed;
- `failed_create`: authentication passed, but article creation failed.

Authentication failures are logged structurally but are not persisted to this audit table in the first version, to avoid attacker-controlled database growth.

## Idempotency

The idempotency key is mandatory and unique.

Behavior:

- If no row exists, create an audit row inside the same transaction as the article creation.
- If a row exists with the same `request_hash` and `status=created`, return the existing article response.
- If a row exists with a different `request_hash`, return `409 Conflict`.
- If a row exists with a failed status and the same `request_hash`, allow a retry only if no article was created.

The request hash is the lowercase hex SHA256 digest of the raw request body.

## Article Creation Rules

Automation-created articles:

- use the earliest admin user by `created_at` as `owner`;
- return `422 Unprocessable Entity` when no admin user exists;
- always set `is_publish=false`;
- set `created_by_automation=true`;
- set `automation_status=pending_review`;
- set `automation_request_id` to the audit row ID;
- calculate `read_time` using the same logic as article updates;
- require `category_id`; no default category fallback is used in this first version.

The creation path should reuse sqlc and transaction patterns already used by article write flows. Cache invalidation is not needed for unpublished drafts.

## Email Notification

Use the existing Asynq worker and mail module.

### Success Notification

Send an owner email after successful draft creation:

```text
Subject: Nostalgia 自动化草稿待审核：<title>

Codex 已创建一篇未发布草稿。
标题：<title>
分类：<category name>
编辑地址：<DOMAIN>/backend/articles/<article_id>
幂等键：<idempotency_key>
生成模型：<generation_model>
```

### Failure Notification

Send failure email only after authentication succeeds and business validation or article creation fails:

```text
Subject: Nostalgia 自动化草稿创建失败

认证已通过，但草稿未创建。
原因：<error_message>
幂等键：<idempotency_key>
标题：<title when available>
```

Authentication failures must not send email.

### Notification Failure

Email delivery failure should be logged and retried by Asynq. Article creation should not be rolled back because notification failed after the article transaction commits.

## Backend UI

The backend article list should display a compact badge for automation-created drafts:

```text
自动化草稿
```

Display rule:

```text
created_by_automation == true
AND automation_status == "pending_review"
AND is_publish == false
```

The article edit page can remain unchanged in the first version. The owner reviews the title, summary, category, slug, cover, and content, then uses the existing publish flow.

When an automation-created article is published, backend update logic should set `automation_status=published`.

## Security And Abuse Controls

- HMAC signing is mandatory.
- Timestamp replay window defaults to five minutes.
- Idempotency key is mandatory.
- Request body size must be bounded.
- Daily draft count limit defaults to one draft per day.
- Authentication failures must use generic error responses.
- Business validation errors may be explicit enough for Codex automation to fix the next request.
- Secrets must only come from runtime environment variables or equivalent secret stores.
- Secrets must never be returned by API responses, stored in audit rows, logged, committed, or exposed to the frontend.

## Error Handling

- Missing automation configuration returns `404 Not Found` so disabled automation is not advertised as an available write surface.
- Authentication failure returns `401 Unauthorized` with a generic message.
- Timestamp outside the allowed window returns `401 Unauthorized`.
- JSON binding failure returns `400 Bad Request`.
- Validation failure returns `422 Unprocessable Entity`.
- Idempotency conflict returns `409 Conflict`.
- Slug conflict returns `409 Conflict`.
- Daily limit conflict returns `409 Conflict`.
- Unexpected database or queue failure returns `500 Internal Server Error`.

## Observability

Use structured logs with fields:

- `module=automation`
- `action=create_article_draft`
- `key_id`
- `idempotency_key`
- `request_hash`
- `article_id`
- `status`
- `client_ip`

Do not log the raw request body, source prompt, content, HMAC secret, or signature.

## Testing Strategy

Backend tests should cover:

- valid HMAC request creates an unpublished automation draft;
- missing signature is rejected;
- invalid signature is rejected;
- expired timestamp is rejected;
- idempotent replay returns the existing article;
- idempotency key with different body returns conflict;
- slug conflict returns conflict;
- missing or invalid category returns validation error;
- daily draft limit returns conflict;
- authenticated validation failure records audit and enqueues failure email;
- authentication failure does not enqueue email;
- backend list response includes automation fields needed by the UI.

Frontend tests should cover:

- backend article list displays `自动化草稿` badge for pending automation drafts;
- normal drafts do not show the badge.

## Rollout

1. Add database migration and sqlc queries.
2. Add config fields and example documentation.
3. Add HMAC verifier with unit tests.
4. Add automation draft API and handler tests.
5. Add notification worker payload and tests.
6. Update backend article list data shape and UI badge.
7. Run full backend and frontend verification.

## Acceptance Criteria

- Codex automation can create a complete unpublished draft through a signed API request.
- The API rejects unsigned, stale, or incorrectly signed requests.
- Retries with the same idempotency key and body are safe.
- Different bodies under the same idempotency key cannot create multiple drafts.
- Created drafts are never published automatically.
- Owner receives email for successful draft creation.
- Authenticated business failures can notify the owner.
- Backend article list clearly marks automation-created pending drafts.
- No automation secret is exposed in logs, API responses, frontend code, or committed files.
