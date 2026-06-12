# AI Polish Editor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an owner-only AI polish workflow for CKEditor article authoring, including backend AI provider configuration, `/v1/ai/polish`, frontend suggestions, and documentation.

**Architecture:** Add a small `internal/ai` package that owns prompt construction, OpenAI-compatible HTTP calls, response normalization, and secret redaction. Expose the feature through the existing gRPC-Gateway admin surface so `adminHttp` can reuse JWT refresh and error handling. Keep frontend application explicit: suggestions are previewed and only change local editor state when the owner applies them.

**Tech Stack:** Go, gRPC, gRPC-Gateway, Protocol Buffers, Viper config, `net/http`, Vue 3, TypeScript, CKEditor 5, Bun tests.

---

## File Map

- `util/config.go`: add AI polish runtime config fields and defaults.
- `util/config_test.go`: prove env loading, defaults, and env key exposure.
- `internal/ai/types.go`: define modes, targets, requests, responses, typed errors, and validation helpers.
- `internal/ai/prompt.go`: build provider prompts from sanitized polish requests.
- `internal/ai/openai_compatible.go`: call `{AI_POLISH_BASE_URL}/chat/completions` and parse strict JSON suggestions.
- `internal/ai/openai_compatible_test.go`: test provider request shape, timeout, parsing, and redaction.
- `proto/rpc_polish_text.proto`: define admin polish request and response messages.
- `proto/service_nostalgia.proto`: register `PolishText` as `POST /v1/ai/polish`.
- `pb/*.go`: generated protobuf and gateway files from `make proto`.
- `gapi/server.go`: hold an `ai.TextPolisher` on the gapi server.
- `gapi/rpc_polish_text.go`: authorize admin, validate input, call AI client, map errors.
- `gapi/rpc_polish_text_test.go`: test auth, invalid input, disabled config, success, provider failure, and secret safety.
- `.github/workflows/test.yml`, `.env.example`, `README.md`, `docker-compose.yaml`, `docker-compose.dev.yaml`: document placeholder AI env names.
- `web/frontend/src/admin/types.ts`: add AI polish request/response types.
- `web/frontend/src/admin/api/adminAiApi.ts`: add `/ai/polish` wrapper.
- `web/frontend/src/admin/ai/polish.ts`: pure frontend helpers for labels, text extraction, truncation, and suggestion application inputs.
- `web/frontend/src/admin/ai/polish.test.ts`: test frontend helpers.
- `web/frontend/src/views/admin/AdminArticleEditorView.vue`: add UI actions and suggestion panel.

## Task 1: Backend Config And AI Client

**Files:**
- Modify: `util/config.go`
- Modify: `util/config_test.go`
- Create: `internal/ai/types.go`
- Create: `internal/ai/prompt.go`
- Create: `internal/ai/openai_compatible.go`
- Create: `internal/ai/openai_compatible_test.go`

- [ ] **Step 1: Write failing config tests**

Add tests in `util/config_test.go`:

```go
func TestLoadConfigAIPolishOverrides(t *testing.T) {
	configPath := t.TempDir() + string(os.PathSeparator)
	setConfigEnv(t, map[string]string{
		"AI_POLISH_PROVIDER":          "openai_compatible",
		"AI_POLISH_BASE_URL":          "https://ai.example.com/v1",
		"AI_POLISH_API_KEY":           "runtime-secret",
		"AI_POLISH_MODEL":             "writer-model",
		"AI_POLISH_TIMEOUT":           "45s",
		"AI_POLISH_MAX_INPUT_CHARS":   "7000",
		"AI_POLISH_MAX_CONTEXT_CHARS": "5000",
		"AI_POLISH_MAX_SUGGESTIONS":   "2",
	})

	config, err := LoadConfig(configPath)
	require.NoError(t, err)
	require.Equal(t, "openai_compatible", config.AIPolishProvider)
	require.Equal(t, "https://ai.example.com/v1", config.AIPolishBaseURL)
	require.Equal(t, "runtime-secret", config.AIPolishAPIKey)
	require.Equal(t, "writer-model", config.AIPolishModel)
	require.Equal(t, 45*time.Second, config.AIPolishTimeout)
	require.Equal(t, 7000, config.AIPolishMaxInputChars)
	require.Equal(t, 5000, config.AIPolishMaxContextChars)
	require.Equal(t, 2, config.AIPolishMaxSuggestions)
}
```

- [ ] **Step 2: Verify config tests fail**

Run: `go test ./util -run 'TestLoadConfigAIPolish|TestConfigExposesAIPolish' -count=1`

Expected: FAIL because `util.Config` does not have AI polish fields yet.

- [ ] **Step 3: Implement config fields and defaults**

Add these fields to `util.Config`:

```go
AIPolishProvider        string        `mapstructure:"AI_POLISH_PROVIDER"`
AIPolishBaseURL         string        `mapstructure:"AI_POLISH_BASE_URL"`
AIPolishAPIKey          string        `mapstructure:"AI_POLISH_API_KEY"`
AIPolishModel           string        `mapstructure:"AI_POLISH_MODEL"`
AIPolishTimeout         time.Duration `mapstructure:"AI_POLISH_TIMEOUT"`
AIPolishMaxInputChars   int           `mapstructure:"AI_POLISH_MAX_INPUT_CHARS"`
AIPolishMaxContextChars int           `mapstructure:"AI_POLISH_MAX_CONTEXT_CHARS"`
AIPolishMaxSuggestions  int           `mapstructure:"AI_POLISH_MAX_SUGGESTIONS"`
```

Add defaults:

```go
configReader.SetDefault("AI_POLISH_PROVIDER", "openai_compatible")
configReader.SetDefault("AI_POLISH_TIMEOUT", 30*time.Second)
configReader.SetDefault("AI_POLISH_MAX_INPUT_CHARS", 6000)
configReader.SetDefault("AI_POLISH_MAX_CONTEXT_CHARS", 4000)
configReader.SetDefault("AI_POLISH_MAX_SUGGESTIONS", 3)
```

- [ ] **Step 4: Add AI client tests**

Create `internal/ai/openai_compatible_test.go` with tests for valid response parsing, upstream failures, malformed JSON, context timeout, and redacted errors. Use `httptest.Server` and assert the request path is `/chat/completions`, the `Authorization` header is present only in the provider request, and returned errors do not contain the test API key.

- [ ] **Step 5: Verify AI client tests fail**

Run: `go test ./internal/ai -count=1`

Expected: FAIL because `internal/ai` implementation does not exist yet.

- [ ] **Step 6: Implement AI package**

Implement:

```go
type TextPolisher interface {
	Polish(ctx context.Context, req PolishRequest) (PolishResponse, error)
}
```

Use typed errors:

```go
var (
	ErrDisabled          = errors.New("ai polish is not configured")
	ErrInvalidInput      = errors.New("invalid ai polish input")
	ErrProviderFailure   = errors.New("ai provider failure")
	ErrMalformedResponse = errors.New("malformed ai provider response")
)
```

`NewOpenAICompatiblePolisher(config util.Config)` should return a disabled polisher when required config is incomplete. The OpenAI-compatible client should post to `strings.TrimRight(baseURL, "/") + "/chat/completions"` with a strict JSON-output instruction and parse suggestions from the first choice content.

- [ ] **Step 7: Run tests and commit**

Run:

```bash
gofmt -w util/config.go util/config_test.go internal/ai
go test ./util ./internal/ai -count=1
```

Expected: PASS.

Commit:

```bash
git add util/config.go util/config_test.go internal/ai
git commit -m "feat: add ai polish client"
```

## Task 2: Admin RPC

**Files:**
- Create: `proto/rpc_polish_text.proto`
- Modify: `proto/service_nostalgia.proto`
- Generate: `pb/*.go`
- Modify: `gapi/server.go`
- Create: `gapi/rpc_polish_text.go`
- Create: `gapi/rpc_polish_text_test.go`

- [ ] **Step 1: Write failing gapi tests**

Create `gapi/rpc_polish_text_test.go` with table tests:

```go
func TestPolishTextRequiresAdmin(t *testing.T) {
	server := newTestServer(t, newGAPITestStore(mockdb.NewMockStore(gomock.NewController(t))), nil, nil)
	ctx := newContextWithUserBearerToken(t, server.tokenMaker, util.RandUserID(), "visitor", util.Visitor, time.Minute)
	_, err := server.PolishText(ctx, &pb.PolishTextRequest{Mode: "improve", Target: "content_selection", Text: "hello"})
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, st.Code())
}
```

Also test invalid mode, disabled config, success with a fake `ai.TextPolisher`, and provider failure mapping to `codes.Unavailable`.

- [ ] **Step 2: Verify gapi tests fail**

Run: `go test ./gapi -run TestPolishText -count=1`

Expected: FAIL because proto messages and `PolishText` do not exist.

- [ ] **Step 3: Add proto and generate**

Create `proto/rpc_polish_text.proto`:

```proto
syntax = "proto3";

package pb;

option go_package = "github.com/MonitorAllen/nostalgia/pb";

message PolishTextRequest {
  string mode = 1;
  string target = 2;
  string text = 3;
  string article_id = 4;
  string article_title = 5;
  string article_summary = 6;
  string article_excerpt = 7;
  string locale = 8;
}

message PolishSuggestion {
  string content = 1;
  string reason = 2;
}

message PolishTextResponse {
  repeated PolishSuggestion suggestions = 1;
  string mode = 2;
  string target = 3;
  string model = 4;
}
```

Import it in `proto/service_nostalgia.proto` and add:

```proto
rpc PolishText (PolishTextRequest) returns (PolishTextResponse) {
  option (google.api.http) = {
    post: "/v1/ai/polish"
    body: "*"
  };
}
```

Run: `make proto`

- [ ] **Step 4: Implement RPC**

Add `textPolisher ai.TextPolisher` to `gapi.Server`, initialize it in `NewServer`, and implement `PolishText`:

```go
func (server *Server) PolishText(ctx context.Context, req *pb.PolishTextRequest) (*pb.PolishTextResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}
	if server.textPolisher == nil {
		return nil, status.Error(codes.FailedPrecondition, "AI 润色尚未配置")
	}
	result, err := server.textPolisher.Polish(ctx, ai.PolishRequest{...})
	if err != nil {
		return nil, mapAIPolishError(err)
	}
	return convertPolishResponse(result), nil
}
```

- [ ] **Step 5: Run tests and commit**

Run:

```bash
gofmt -w gapi internal/ai
go test ./gapi ./internal/ai -count=1
```

Expected: PASS.

Commit:

```bash
git add proto pb gapi
git commit -m "feat: expose ai polish admin rpc"
```

## Task 3: Frontend API And Helpers

**Files:**
- Modify: `web/frontend/src/admin/types.ts`
- Create: `web/frontend/src/admin/api/adminAiApi.ts`
- Create: `web/frontend/src/admin/ai/polish.ts`
- Create: `web/frontend/src/admin/ai/polish.test.ts`

- [ ] **Step 1: Write failing frontend helper tests**

Create `web/frontend/src/admin/ai/polish.test.ts`:

```ts
import { describe, expect, test } from 'bun:test'
import { buildAIPolishRequest, getAIPolishModeLabel, normalizeSelectedText, truncateForAIPolish } from './polish'

describe('AI polish helpers', () => {
  test('normalizes selected text', () => {
    expect(normalizeSelectedText('  hello\\n\\n world  ')).toBe('hello\\nworld')
  })

  test('truncates long context', () => {
    expect(truncateForAIPolish('abcdef', 4)).toBe('abcd')
  })

  test('builds content selection requests', () => {
    expect(buildAIPolishRequest({
      mode: 'improve',
      target: 'content_selection',
      text: 'hello',
      articleTitle: 'Title',
      articleSummary: 'Summary',
      articleExcerpt: 'Excerpt',
    })).toEqual({
      mode: 'improve',
      target: 'content_selection',
      text: 'hello',
      article_title: 'Title',
      article_summary: 'Summary',
      article_excerpt: 'Excerpt',
      locale: 'zh-CN',
    })
  })

  test('labels modes', () => {
    expect(getAIPolishModeLabel('shorten')).toBe('精简')
  })
})
```

- [ ] **Step 2: Verify frontend tests fail**

Run: `cd web/frontend && bun test src/admin/ai/polish.test.ts`

Expected: FAIL because helper module does not exist.

- [ ] **Step 3: Implement frontend types, API, and helpers**

Add request and response types in `types.ts`, create `adminAiApi.ts`:

```ts
export function polishAdminText(data: AdminAIPolishRequest) {
  return adminHttp.post<AdminAIPolishResponse>('/ai/polish', data)
}
```

Implement helper functions with explicit modes and labels.

- [ ] **Step 4: Run tests and commit**

Run:

```bash
cd web/frontend && bun test src/admin/ai/polish.test.ts
cd web/frontend && bun run type-check
```

Expected: PASS.

Commit:

```bash
git add web/frontend/src/admin
git commit -m "feat: add admin ai polish client"
```

## Task 4: Editor UI And Documentation

**Files:**
- Modify: `web/frontend/src/views/admin/AdminArticleEditorView.vue`
- Modify: `.env.example`
- Modify: `README.md`
- Modify: `docker-compose.yaml`
- Modify: `docker-compose.dev.yaml`
- Modify: `.github/workflows/test.yml`

- [ ] **Step 1: Add editor UI state and handlers**

In `AdminArticleEditorView.vue`, store the editor instance, selected text, polish mode, suggestions, loading state, and active target. Use CKEditor model selection to read selected plain text. Use the helper functions from `admin/ai/polish.ts` and call `polishAdminText`.

- [ ] **Step 2: Add editor controls**

Add a `Sparkles` icon button near save controls for selected content and compact title/summary candidate buttons. Add a non-nested suggestion panel near the editor/sidebar with `替换`, `插入`, `复制`, and `取消`.

- [ ] **Step 3: Add environment documentation**

Add placeholder AI env names to `.env.example`, README env section, Compose API env blocks, and CI test env. Never add a real key.

- [ ] **Step 4: Run frontend and config verification**

Run:

```bash
cd web/frontend && bun test
cd web/frontend && bun run type-check
cd web/frontend && bun run build
docker compose config --quiet
docker compose -f docker-compose.dev.yaml config --quiet
git diff --check
```

Expected: PASS. Existing Vite chunk warning is acceptable.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/views/admin/AdminArticleEditorView.vue .env.example README.md docker-compose.yaml docker-compose.dev.yaml .github/workflows/test.yml
git commit -m "feat: integrate ai polish editor UI"
```

## Task 5: Final Verification

**Files:**
- Verify all changed files.

- [ ] **Step 1: Run full verification**

Run:

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

Expected: all commands exit 0. Vite large chunk warning may appear and is acceptable.

- [ ] **Step 2: Inspect final diff**

Run:

```bash
git status --short --branch
git log --oneline -8 --decorate
```

Expected: branch `feature/ai-polish-editor` contains separate docs/config/backend/frontend commits and no unstaged changes after commits.
