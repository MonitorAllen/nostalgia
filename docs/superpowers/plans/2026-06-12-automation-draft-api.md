# Automation Draft API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a signed automation API that lets Codex create unpublished article drafts, notifies the owner by email, and marks the draft in the backend list for manual review.

**Architecture:** Add a dedicated Gin endpoint under `/api/automation/articles/drafts` with HMAC authentication, timestamp validation, idempotency, and daily draft limits. Store audit data in PostgreSQL through sqlc and create the article in a transaction. Reuse Asynq and the mail sender for owner notifications, and extend the backend list data shape so Vue can show an `自动化草稿` badge.

**Tech Stack:** Go, Gin, PostgreSQL, sqlc, golang-migrate, GoMock, Asynq, existing mail module, Protocol Buffers/gRPC-Gateway, Vue 3, TypeScript, Bun.

---

## File Structure

- Create `internal/automation/signature.go`: HMAC request signature helpers.
- Create `internal/automation/signature_test.go`: verifier unit tests.
- Modify `util/config.go` and `util/config_test.go`: automation configuration fields and defaults.
- Add `db/migration/000010_add_automation_drafts.up.sql` and `.down.sql`: article automation fields and audit table.
- Modify `db/query/article.sql`: automation article insert, admin list fields, publish status update.
- Add `db/query/automation.sql`: audit/idempotency queries and daily limit query.
- Modify `db/query/user.sql`: query earliest admin owner.
- Add `db/sqlc/tx_create_automation_article.go`: transaction for audit row plus article creation.
- Regenerate `db/sqlc/*` with `make sqlc`.
- Regenerate mocks with `make mock`.
- Modify `proto/article.proto`: add automation fields to admin article payload.
- Run `make proto` after proto changes.
- Create `worker/task_notify_automation_draft.go`: success/failure notification payload, distributor, processor.
- Modify `worker/distributor.go` and `worker/processor.go`: register notification task.
- Create `api/automation.go`: HTTP handler for signed draft creation.
- Add `api/automation_test.go`: handler tests for auth, idempotency, validation, limits, and notification behavior.
- Modify `api/server.go`: register automation route.
- Modify `gapi/converter.go`, `gapi/rpc_list_articles.go`, and related tests: expose automation fields in admin article list.
- Modify `web/frontend/src/admin/types.ts`: add automation fields.
- Modify `web/frontend/src/views/admin/AdminArticleListView.vue`: show automation draft badge.
- Add or modify frontend tests for badge behavior.
- Update `.env.example`, README, Docker Compose, and CI env with automation variable names.

## Task 1: Config And HMAC Verifier

**Files:**
- Modify: `util/config.go`
- Modify: `util/config_test.go`
- Create: `internal/automation/signature.go`
- Create: `internal/automation/signature_test.go`

- [x] **Step 1: Write failing config tests**

Add tests that expect:

```go
require.Equal(t, "codex-daily-writer", config.AutomationHMACKeyID)
require.Equal(t, "secret", config.AutomationHMACSecret)
require.Equal(t, 5*time.Minute, config.AutomationSignatureTTL)
require.Equal(t, int64(1), config.AutomationDailyDraftLimit)
require.Equal(t, "owner@example.com", config.AutomationNotifyEmail)
```

Run:

```bash
go test ./util -run 'TestLoadConfigAutomation|TestConfigExposesAutomation'
```

Expected: fail because automation config fields do not exist.

- [x] **Step 2: Implement config fields**

Add fields to `util.Config`:

```go
AutomationHMACKeyID       string        `mapstructure:"AUTOMATION_HMAC_KEY_ID"`
AutomationHMACSecret      string        `mapstructure:"AUTOMATION_HMAC_SECRET"`
AutomationSignatureTTL    time.Duration `mapstructure:"AUTOMATION_SIGNATURE_TTL"`
AutomationDailyDraftLimit int64         `mapstructure:"AUTOMATION_DAILY_DRAFT_LIMIT"`
AutomationNotifyEmail     string        `mapstructure:"AUTOMATION_NOTIFY_EMAIL"`
```

Set defaults in `LoadConfig`:

```go
configReader.SetDefault("AUTOMATION_SIGNATURE_TTL", 5*time.Minute)
configReader.SetDefault("AUTOMATION_DAILY_DRAFT_LIMIT", 1)
```

- [x] **Step 3: Write failing HMAC verifier tests**

Cover:

- valid signature is accepted;
- invalid signature is rejected;
- stale timestamp is rejected;
- missing required header is rejected;
- base string includes method, path, timestamp, idempotency key, and body hash.

Run:

```bash
go test ./internal/automation
```

Expected: fail because package does not exist.

- [x] **Step 4: Implement `internal/automation` verifier**

Expose:

```go
type SignatureInput struct {
	Method         string
	Path           string
	Timestamp      string
	IdempotencyKey string
	Body           []byte
	Now            time.Time
	TTL            time.Duration
	KeyID          string
	ExpectedKeyID  string
	Secret         string
	Signature      string
}

func SHA256Hex(body []byte) string
func SignatureBaseString(method, path, timestamp, idempotencyKey, bodyHash string) string
func Sign(secret, base string) string
func VerifySignature(input SignatureInput) error
```

Use `hmac.Equal` for comparison and require `v1=` signature prefix.

- [x] **Step 5: Verify and commit**

Run:

```bash
gofmt -w util/config.go util/config_test.go internal/automation/*.go
go test ./util ./internal/automation
git add util/config.go util/config_test.go internal/automation
git commit -m "feat: add automation signature config"
```

## Task 2: Database Model, Queries, And Generated Code

**Files:**
- Create: `db/migration/000010_add_automation_drafts.up.sql`
- Create: `db/migration/000010_add_automation_drafts.down.sql`
- Modify: `db/query/article.sql`
- Add: `db/query/automation.sql`
- Modify: `db/query/user.sql`
- Create: `db/sqlc/tx_create_automation_article.go`
- Modify generated: `db/sqlc/*`
- Modify generated: `db/mock/store.go`

- [x] **Step 1: Write failing database tests**

Add tests in `db/sqlc/automation_test.go` that:

- create an automation article request row;
- create an automation draft article in a transaction;
- replay the same idempotency key and request hash;
- count today's automation drafts;
- fetch the earliest admin user as owner.

Run:

```bash
go test ./db/sqlc -run 'TestAutomation|TestCreateAutomationArticleTx'
```

Expected: fail because tables and queries do not exist.

- [x] **Step 2: Add migration**

`000010_add_automation_drafts.up.sql` adds:

```sql
ALTER TABLE articles
  ADD COLUMN created_by_automation boolean NOT NULL DEFAULT false,
  ADD COLUMN automation_status varchar(32) NOT NULL DEFAULT '',
  ADD COLUMN automation_request_id bigint;

CREATE TABLE automation_article_requests (
  id bigserial PRIMARY KEY,
  idempotency_key varchar(160) UNIQUE NOT NULL,
  request_hash varchar(64) NOT NULL,
  key_id varchar(100) NOT NULL,
  status varchar(32) NOT NULL,
  article_id uuid,
  title varchar NOT NULL DEFAULT '',
  source_topic varchar NOT NULL DEFAULT '',
  source_prompt text NOT NULL DEFAULT '',
  generation_model varchar NOT NULL DEFAULT '',
  error_message text NOT NULL DEFAULT '',
  client_ip varchar NOT NULL DEFAULT '',
  user_agent varchar NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

ALTER TABLE articles
  ADD CONSTRAINT articles_automation_request_id_fkey
  FOREIGN KEY (automation_request_id) REFERENCES automation_article_requests(id);
```

The down migration drops the foreign key, drops the three article columns, and drops `automation_article_requests`.

- [x] **Step 3: Add sqlc queries**

Add:

```sql
-- name: GetFirstAdminUser :one
SELECT * FROM users WHERE role = 'admin' ORDER BY created_at ASC LIMIT 1;
```

Add automation audit queries:

```sql
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
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetAutomationArticleRequestByIdempotencyKey :one
SELECT * FROM automation_article_requests WHERE idempotency_key = $1 LIMIT 1;

-- name: MarkAutomationArticleRequestCreated :one
UPDATE automation_article_requests
SET status = 'created', article_id = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CountAutomationDraftsToday :one
SELECT count(*)
FROM automation_article_requests
WHERE status = 'created'
  AND created_at >= date_trunc('day', now());
```

Add an article insert query that includes automation fields:

```sql
-- name: CreateAutomationArticle :one
INSERT INTO articles (
  id,
  title,
  summary,
  content,
  is_publish,
  owner,
  category_id,
  cover,
  slug,
  check_outdated,
  last_updated,
  read_time,
  created_by_automation,
  automation_status,
  automation_request_id
) VALUES (
  $1, $2, $3, $4, false, $5, $6, $7, $8, $9, now(), $10, true, 'pending_review', $11
)
RETURNING *;
```

Extend `ListAllArticles` SELECT with `created_by_automation`, `automation_status`, and `automation_request_id`.

- [x] **Step 4: Generate sqlc and mocks**

Run:

```bash
make sqlc
make mock
```

- [x] **Step 5: Implement transaction**

Add `CreateAutomationArticleTx(ctx, arg)` to `db.Store` and `SQLStore`. It should create the audit row, create the article, mark the request created, and return both rows.

- [x] **Step 6: Verify and commit**

Run:

```bash
gofmt -w db/sqlc db/mock
go test ./db/sqlc
git add db/migration db/query db/sqlc db/mock
git commit -m "feat: add automation draft persistence"
```

## Task 3: Automation Email Notification Task

**Files:**
- Create: `worker/task_notify_automation_draft.go`
- Modify: `worker/distributor.go`
- Modify: `worker/processor.go`
- Modify generated: `worker/mock/distributor.go`

- [x] **Step 1: Write failing worker tests**

Add `worker/task_notify_automation_draft_test.go` to verify:

- success payload sends a review email;
- failure payload sends a failure email;
- invalid payload skips retry.

Run:

```bash
go test ./worker -run TestProcessTaskNotifyAutomationDraft
```

Expected: fail because task does not exist.

- [x] **Step 2: Implement task payload and distributor**

Add:

```go
const TaskNotifyAutomationDraft = "task:notify_automation_draft"

type PayloadNotifyAutomationDraft struct {
	Kind            string    `json:"kind"`
	ArticleID       uuid.UUID `json:"article_id"`
	Title           string    `json:"title"`
	CategoryName    string    `json:"category_name"`
	ReviewURL       string    `json:"review_url"`
	IdempotencyKey  string    `json:"idempotency_key"`
	GenerationModel string    `json:"generation_model"`
	ErrorMessage    string    `json:"error_message"`
	NotifyEmail     string    `json:"notify_email"`
}
```

Add `DistributeTaskNotifyAutomationDraft` to `TaskDistributor`.

- [x] **Step 3: Implement processor**

Register the task in `processor.Start()`. Build email subject/content based on `Kind == "success"` or `Kind == "failure"`.

- [x] **Step 4: Regenerate mocks, verify, commit**

Run:

```bash
make mock
gofmt -w worker
go test ./worker
git add worker
git commit -m "feat: notify automation draft review"
```

## Task 4: Automation Draft API

**Files:**
- Create: `api/automation.go`
- Create: `api/automation_test.go`
- Modify: `api/server.go`
- Modify: `gapi/rpc_update_article.go`
- Modify: `gapi/rpc_update_article_test.go`

- [x] **Step 1: Write failing handler tests**

Add tests that cover:

- valid signed request creates unpublished automation draft;
- unsigned request returns `401`;
- invalid signature returns `401`;
- stale timestamp returns `401`;
- same idempotency key and same body returns replay response;
- same idempotency key and different body returns `409`;
- missing category returns `422`;
- slug conflict returns `409`;
- daily limit returns `409`;
- authenticated business validation failure enqueues failure email;
- authentication failure does not enqueue email.

Run:

```bash
go test ./api -run TestCreateAutomationArticleDraft
```

Expected: fail because the route and handler do not exist.

- [x] **Step 2: Register route**

In `api/server.go`:

```go
public.POST("/automation/articles/drafts", server.createAutomationArticleDraft)
```

- [x] **Step 3: Implement handler**

Handler responsibilities:

- return `404` when automation config is missing;
- read and restore raw body;
- validate HMAC with `internal/automation`;
- parse JSON body;
- check idempotency row;
- enforce daily draft limit;
- fetch category and first admin owner;
- create article through `CreateAutomationArticleTx`;
- enqueue success notification;
- record authenticated validation failures and enqueue failure notification;
- return `201` for new draft and `200` for replay.

- [x] **Step 4: Publish status update**

When an automation-created article is published through existing admin update flow, set `automation_status='published'`.

- [x] **Step 5: Verify and commit**

Run:

```bash
gofmt -w api gapi
go test ./api ./gapi
git add api gapi
git commit -m "feat: add signed automation draft endpoint"
```

## Task 5: Admin List Automation Badge

**Files:**
- Modify: `proto/article.proto`
- Modify generated: `pb/*`
- Modify: `gapi/converter.go`
- Modify: `web/frontend/src/admin/types.ts`
- Modify: `web/frontend/src/views/admin/AdminArticleListView.vue`
- Add or modify frontend tests under `web/frontend/src/admin` or `web/frontend/src/views/admin`

- [x] **Step 1: Write failing frontend test**

Add a test that renders or inspects the list view behavior for:

```ts
{
  created_by_automation: true,
  automation_status: 'pending_review',
  is_publish: false
}
```

Expected: `自动化草稿` appears only for that state.

Run:

```bash
cd web/frontend
bun test
```

- [x] **Step 2: Extend proto and regenerate**

Add to `Article`:

```proto
optional bool created_by_automation = 18;
string automation_status = 19;
```

Run:

```bash
make proto
```

- [x] **Step 3: Populate admin list fields**

Update `convertArticle` and related converters to include automation fields from sqlc rows.

- [x] **Step 4: Update frontend types and badge**

Add:

```ts
created_by_automation?: boolean
automation_status?: string
```

Add helper:

```ts
const isAutomationDraft = (article: AdminArticle) =>
  Boolean(article.created_by_automation) &&
  article.automation_status === 'pending_review' &&
  !article.is_publish
```

Render `<AppBadge tone="warning">自动化草稿</AppBadge>` near the existing draft/published badge.

- [x] **Step 5: Verify and commit**

Run:

```bash
go test ./gapi
cd web/frontend && bun test && bun run type-check && bun run build
git add proto pb gapi web/frontend/src
git commit -m "feat: mark automation drafts in admin"
```

## Task 6: Documentation, Config Examples, And Full Verification

**Files:**
- Modify: `.env.example`
- Modify: `.github/workflows/test.yml`
- Modify: `docker-compose.yaml`
- Modify: `docker-compose.dev.yaml`
- Modify: `README.md`
- Modify: `docs/superpowers/plans/2026-06-12-automation-draft-api.md`

- [x] **Step 1: Update config docs**

Add placeholder variables:

```env
AUTOMATION_HMAC_KEY_ID=codex-daily-writer
AUTOMATION_HMAC_SECRET=replace-with-a-random-secret
AUTOMATION_SIGNATURE_TTL=5m
AUTOMATION_DAILY_DRAFT_LIMIT=1
AUTOMATION_NOTIFY_EMAIL=owner@example.com
```

- [x] **Step 2: Run full verification**

Run:

```bash
go test -v -cover -short -count=1 ./...
cd web/frontend && bun test && bun run type-check && bun run build
docker compose config --quiet
docker compose -f docker-compose.dev.yaml config --quiet
git diff --check
```

- [x] **Step 3: Mark plan complete and commit**

Mark completed plan checkboxes, then:

```bash
git add .env.example .github/workflows/test.yml docker-compose.yaml docker-compose.dev.yaml README.md docs/superpowers/plans/2026-06-12-automation-draft-api.md
git commit -m "docs: document automation draft api configuration"
```

## Self-Review

- Spec coverage: this plan covers HMAC auth, timestamp replay protection, idempotency, audit persistence, daily limit, unpublished draft creation, owner email notification, backend list marker, docs, and verification.
- Placeholder scan: no open-ended placeholder tasks remain; exact file paths, commands, and expected behavior are listed.
- Type consistency: names align around `AutomationArticleRequest`, `CreateAutomationArticleTx`, `PayloadNotifyAutomationDraft`, `created_by_automation`, and `automation_status`.
