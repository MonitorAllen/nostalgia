# AI Polish Editor Design

## Context

Nostalgia now has a secure automation draft API for Codex-created drafts, but owner-side writing still depends on manual editing inside the backend CKEditor. The next useful step is to add an owner-only AI polish workflow that helps refine selected article text, titles, and summaries without publishing or saving anything automatically.

The current admin editor already has the right foundation:

- `/backend` lives inside the unified Vue frontend.
- Admin API calls go through `web/frontend/src/admin/api/adminHttp.ts` with `/v1` as the base URL.
- `/v1` is served by gRPC-Gateway and guarded by admin JWT checks in `gapi.authorizeAdmin`.
- `AdminArticleEditorView.vue` already tracks dirty state, stores unsaved drafts in `sessionStorage`, shows save status, and uses toast feedback.
- Auth secrets and other sensitive runtime values already belong in backend configuration and deployment secrets, not frontend code.

This design adds AI polish as an admin authoring assistant. It is not part of public reading pages and it does not give AI any ability to publish content.

## Goals

- Let the owner polish selected CKEditor article text from the backend editor.
- Let the owner request title and summary candidates from the current article context.
- Keep all AI provider secrets server-side.
- Reuse the existing `/v1` admin API surface, JWT admin authorization, frontend `adminHttp`, and toast patterns.
- Return suggestions for owner review instead of modifying or saving articles automatically.
- Keep the first version provider-neutral through an OpenAI-compatible HTTP client.
- Make later SEO/GEO generation able to reuse the same AI client, configuration, and prompt boundary.

## Non-Goals

- Do not expose AI API keys, provider secrets, prompts with secrets, or raw provider responses to the browser.
- Do not add a database-backed AI settings UI in this phase.
- Do not store polish prompts, article selections, or generated suggestions in PostgreSQL.
- Do not stream model output in this phase.
- Do not let AI create, update, publish, or delete articles directly through this feature.
- Do not replace CKEditor or redesign the whole editor shell.
- Do not implement SEO/GEO metadata generation in this branch.
- Do not add multi-user billing, usage analytics, or a full prompt management system.

## Approach Decision

### Recommended: Backend `/v1` AI RPC plus runtime secret config

Add a new owner-only gRPC-Gateway endpoint under `/v1/ai/polish`. The backend reads AI provider configuration from runtime environment variables. The frontend calls the endpoint through `adminHttp`, shows suggestions in the editor, and applies a suggestion only after the owner clicks an explicit action.

This is the best first version because it matches the existing admin architecture, avoids secret leakage, and creates a reusable backend AI boundary for future SEO/GEO work.

### Alternative: Browser calls AI provider directly

This would be faster to prototype but is not acceptable because API keys would need to reach the browser or a public token exchange flow. It also makes provider failure handling and request limits harder to control.

### Alternative: Database-backed backend AI settings page

This would satisfy in-app configuration, but it needs encrypted secret storage, key rotation rules, masking behavior, and audit handling. It is useful later, but it is too much surface area for the first polish feature. The first version should use environment variables or Cloudflare Secrets Store injection for secrets.

## High-Level Flow

```text
Owner selects text in CKEditor or chooses title/summary target
  -> Vue calls POST /v1/ai/polish through adminHttp
  -> gRPC-Gateway maps request to gapi.PolishText
  -> authorizeAdmin validates JWT role=admin
  -> validate mode, target, text length, and context length
  -> internal AI client calls configured OpenAI-compatible provider
  -> backend normalizes provider output into suggestions
  -> frontend previews suggestions
  -> owner explicitly replaces/inserts/copies text
  -> existing dirty-state and draft cache detect the local edit
```

No step automatically saves the article. Publishing remains the existing explicit article save flow.

## Backend API Contract

### Endpoint

```http
POST /v1/ai/polish
Authorization: Bearer <admin-access-token>
Content-Type: application/json
```

This endpoint belongs to the admin gRPC-Gateway surface. It must not be exposed under public `/api`.

### Request

```json
{
  "mode": "improve",
  "target": "content_selection",
  "text": "需要润色的选中文本",
  "article_id": "optional-uuid",
  "article_title": "当前文章标题",
  "article_summary": "当前文章摘要",
  "article_excerpt": "当前文章正文的有限长度摘录",
  "locale": "zh-CN"
}
```

Fields:

- `mode`: required. Allowed values are `improve`, `shorten`, `expand`, `title_candidates`, and `summary_candidates`.
- `target`: required. Allowed values are `content_selection`, `title`, and `summary`.
- `text`: required for `content_selection`. For title and summary candidates it may contain the current title or summary, but context can also come from `article_summary` and `article_excerpt`.
- `article_id`: optional article UUID used only for context and logs. The endpoint does not update the article.
- `article_title`: optional bounded context.
- `article_summary`: optional bounded context.
- `article_excerpt`: optional bounded content context for title and summary candidates.
- `locale`: optional. Defaults to `zh-CN`.

Mode and target must be compatible:

- `improve`, `shorten`, and `expand` use `target = content_selection`.
- `title_candidates` uses `target = title`.
- `summary_candidates` uses `target = summary`.

### Response

```json
{
  "suggestions": [
    {
      "content": "润色后的候选文本",
      "reason": "更简洁，语气更自然"
    }
  ],
  "mode": "improve",
  "target": "content_selection",
  "model": "configured-model-name"
}
```

Response rules:

- Return one to three suggestions.
- Do not return provider API keys or provider request headers.
- `model` may expose the configured model name because it is not a secret.
- Suggestions are plain text in the first version. Rich HTML preservation can be a later branch.

### Error Behavior

- Missing or invalid admin token: existing admin auth behavior.
- Visitor token: `PermissionDenied`.
- AI config disabled or incomplete: `FailedPrecondition` with a friendly message.
- Invalid mode, target, empty text, or oversized input: `InvalidArgument`.
- Provider timeout or upstream failure: `Unavailable`.
- Malformed provider response: `Internal`.

Frontend should show a toast and keep the editor unchanged for every failure.

## Configuration

Add backend runtime configuration fields:

```env
AI_POLISH_PROVIDER=openai_compatible
AI_POLISH_BASE_URL=https://api.example.com/v1
AI_POLISH_API_KEY=replace-with-runtime-secret
AI_POLISH_MODEL=replace-with-model-name
AI_POLISH_TIMEOUT=30s
AI_POLISH_MAX_INPUT_CHARS=6000
AI_POLISH_MAX_CONTEXT_CHARS=4000
AI_POLISH_MAX_SUGGESTIONS=3
```

Behavior:

- AI polish is disabled when `AI_POLISH_API_KEY`, `AI_POLISH_BASE_URL`, or `AI_POLISH_MODEL` is empty.
- `AI_POLISH_PROVIDER` defaults to `openai_compatible`.
- `AI_POLISH_TIMEOUT` defaults to `30s`.
- `AI_POLISH_MAX_INPUT_CHARS` defaults to `6000`.
- `AI_POLISH_MAX_CONTEXT_CHARS` defaults to `4000`.
- `AI_POLISH_MAX_SUGGESTIONS` defaults to `3`.
- `HTTP_PROXY_ADDR` may be reused by the AI HTTP client when configured.
- `.env.example`, README, Docker Compose examples, and CI env should document variable names with placeholder values only.
- Production secrets should come from deployment-time secret injection, for example Cloudflare Secrets Store or equivalent GitHub/host secrets.

## AI Client Boundary

Create an internal package such as `internal/ai` with a small interface:

```go
type TextPolisher interface {
    Polish(ctx context.Context, req PolishRequest) (PolishResponse, error)
}
```

Responsibilities:

- Build provider requests from validated backend input.
- Apply request timeout and context cancellation.
- Reuse `HTTP_PROXY_ADDR` when set.
- Redact `Authorization` and API key values from logs and errors.
- Normalize provider responses into `[]Suggestion`.
- Return typed errors that gapi can map to gRPC status codes.

The first provider implementation should target an OpenAI-compatible chat completions style API by posting to `{AI_POLISH_BASE_URL}/chat/completions`. The code should keep the provider behind an interface so a future Responses API, local model, or custom provider can be added without changing CKEditor integration.

## Prompt Policy

The backend owns all system prompts. The browser sends user-selected content and bounded article context, but it does not send arbitrary system instructions.

Prompt requirements:

- Output must preserve the original language unless the mode explicitly asks otherwise. First version always keeps Chinese or mixed technical Chinese/English.
- Keep technical terms, code identifiers, commands, URLs, and quoted code unchanged unless grammar requires surrounding prose changes.
- Do not invent facts, benchmark numbers, dependency versions, citations, or links.
- For `shorten`, reduce wording without removing essential meaning.
- For `expand`, add clarity and transitions but do not add unsupported claims.
- For title and summary candidates, generate concise options based only on provided title, summary, and selected article context.
- Return strict JSON that the backend can parse deterministically.

## Frontend UX

### Editor Entry Points

Add a compact AI polish control to `AdminArticleEditorView.vue`:

- A `Sparkles` icon button near the save controls for selected content polishing.
- Small actions near title and summary controls for title/summary candidates.
- Disable controls while a polish request is in flight.

The text shown in the UI should be short and operational:

- `润色`
- `精简`
- `扩写`
- `标题候选`
- `摘要候选`
- `替换`
- `插入`
- `复制`
- `取消`

### Selected Content Flow

When the owner selects content in CKEditor and clicks AI polish:

1. Frontend reads the current selected text.
2. If no meaningful selection exists, show a warning toast and do not call the API.
3. Show a small panel or dialog with mode choices.
4. Send selected text plus bounded article title/summary context.
5. Show returned suggestions.
6. Owner can replace the selected text, insert at cursor, copy the suggestion, or cancel.

Applying a suggestion changes local editor content only. The existing draft cache and dirty-state watcher should mark the article as modified.

### Title and Summary Candidate Flow

Title and summary actions use current article fields as context:

- Title candidates can be generated from current title, summary, and a bounded excerpt of article content.
- Summary candidates can be generated from title, current summary, and a bounded excerpt of article content.
- Applying a candidate updates the local title or summary field only.
- The owner must still click `保存文章`.

### Loading and Failure States

- While loading, show a small busy state in the polish panel and disable duplicate submits.
- On success, keep suggestions visible until the owner applies or closes them.
- On failure, show a toast with a friendly message and keep existing editor content unchanged.
- If the backend says AI polish is not configured, show a clear message such as `AI 润色尚未配置`.

## Data and Privacy

- Do not store prompts or suggestions in the database in this phase.
- Do not include full article content unless needed for title/summary context, and even then cap context length.
- Do not log selected text, full prompts, provider responses, or generated suggestions.
- Logs may include safe metadata such as mode, target, configured provider name, model name, input length, duration, and status.
- API keys must never appear in frontend bundles, responses, logs, docs, commits, or test fixtures.

## Security

- Only `role = admin` may call `/v1/ai/polish`.
- The endpoint must use existing admin JWT validation, including refresh behavior handled by `adminHttp`.
- The endpoint should fail closed when configuration is incomplete.
- Request size and text length limits must be enforced before provider calls.
- Provider calls must have a timeout.
- The frontend must not trust AI output as HTML in the first version; suggestions are plain text.
- The backend should avoid returning raw upstream error bodies because provider errors may include request fragments.

## Files Expected To Change During Implementation

Backend:

- `util/config.go` and `util/config_test.go`
- `.env.example`, `README.md`, Docker Compose env examples, and CI env
- `proto/rpc_polish_text.proto`
- `proto/service_nostalgia.proto`
- generated `pb/` files after `make proto`
- `gapi/rpc_polish_text.go`
- `gapi/rpc_polish_text_test.go`
- `internal/ai/*`

Frontend:

- `web/frontend/src/admin/types.ts`
- `web/frontend/src/admin/api/adminAiApi.ts`
- `web/frontend/src/admin/ai/*` for pure helpers and tests
- `web/frontend/src/views/admin/AdminArticleEditorView.vue`

No database migration is expected for the first version.

## Testing Plan

Backend:

- Config tests for AI polish env keys, defaults, and disabled behavior.
- Internal AI client tests using `httptest.Server`:
  - sends authorization only to provider;
  - respects timeout;
  - parses successful suggestions;
  - maps malformed response and upstream failures.
- gapi tests:
  - unauthenticated request is rejected;
  - visitor role is rejected;
  - disabled config returns `FailedPrecondition`;
  - invalid mode/target/text returns `InvalidArgument`;
  - successful request returns normalized suggestions;
  - provider failure maps to `Unavailable`;
  - secrets are not included in errors.

Frontend:

- API wrapper test covering the `/ai/polish` request payload and response typing.
- Pure helper tests for:
  - selection text normalization;
  - mode labels;
  - applying title and summary suggestions to local state.
- Type-check the editor integration.

Minimum verification:

```bash
make proto
make test
cd web/frontend && bun test
cd web/frontend && bun run type-check
cd web/frontend && bun run build
docker compose config --quiet
docker compose -f docker-compose.dev.yaml config --quiet
git diff --check
```

The existing Vite large chunk warning remains acceptable unless this branch makes it materially worse.

## Rollout

1. Merge backend config and internal AI client with tests.
2. Add `/v1/ai/polish` RPC and generated protobuf files.
3. Add frontend API wrapper and pure helper tests.
4. Add CKEditor polish panel and title/summary candidate actions.
5. Document env variables and deployment secret expectations.
6. Verify locally and through PR CI.

## Success Criteria

- Admin owner can select article text, request polish suggestions, and apply one without auto-saving.
- Admin owner can request title and summary candidates and apply one locally.
- Visitor and unauthenticated users cannot access the polish endpoint.
- AI provider secrets remain server-side and are never exposed to the frontend.
- Incomplete AI configuration disables the feature safely with a clear message.
- Existing editor save, draft recovery, upload, and leave-guard behavior continue to work.
- Future SEO/GEO work can reuse the AI client and configuration without changing the editor-specific UI.
