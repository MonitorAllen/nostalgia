# AI Polish Immersive Authoring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current brittle AI polish implementation with a stable owner-only writing assistant that uses official SDK adapters, fully custom prompt templates, JSON-first/raw-fallback parsing, and an explicit right-side candidate drawer in the admin editor.

**Architecture:** Keep `/v1/ai/*` as the admin-only API surface. Refactor `internal/ai` into a small `PolishService` that renders prompt templates, selects an official SDK adapter, parses provider output, and returns normalized candidates. Persist prompt templates with the existing AI provider config, then expose them to the AI settings page and the editor candidate drawer.

**Tech Stack:** Go 1.24, gRPC, gRPC-Gateway, sqlc, PostgreSQL JSONB, OpenAI Go SDK `github.com/openai/openai-go/v3`, Anthropic Go SDK `github.com/anthropics/anthropic-sdk-go`, Vue 3, TypeScript, CKEditor 5, Bun.

---

## Scope Check

This is one implementation stream because the backend adapter refactor, prompt persistence, settings UI, and editor candidate drawer all meet at the same `/v1/ai/polish` contract. The work is split into small commits so backend API changes land before the UI relies on them.

## File Map

- `go.mod`, `go.sum`: add official OpenAI and Anthropic SDKs.
- `db/migration/000013_add_ai_prompt_templates.up.sql`: add `prompt_templates jsonb`.
- `db/migration/000013_add_ai_prompt_templates.down.sql`: remove the column.
- `db/query/ai_provider_config.sql`: read and upsert prompt template JSON.
- `db/sqlc/*.go`, `db/mock/store.go`: regenerated via `make sqlc` and `make mock`.
- `proto/rpc_polish_text.proto`: add prompt template fields to get/update AI config responses.
- `pb/*.go`: regenerated via `make proto`.
- `internal/ai/types.go`: keep public modes and add provider adapter/service config types.
- `internal/ai/templates.go`: built-in templates, normalization, and rendering.
- `internal/ai/parser.go`: JSON-first/raw-fallback response parsing.
- `internal/ai/service.go`: validate polish requests, render prompt, call adapter, parse output.
- `internal/ai/openai_adapter.go`: OpenAI Chat Completions and Responses adapter using `openai-go/v3`.
- `internal/ai/anthropic_adapter.go`: Anthropic Messages adapter using `anthropic-sdk-go`.
- `internal/ai/*_test.go`: focused unit tests for templates, parsing, service adapter selection, and SDK request shape.
- `gapi/ai_config.go`: include prompt templates in resolved config, validation, persistence, and responses.
- `gapi/rpc_polish_text.go`: call `ai.NewPolishService` instead of the old combined polisher.
- `gapi/rpc_list_ai_models.go`: keep OpenAI model listing; return unsupported for Anthropic messages in this phase.
- `gapi/rpc_polish_text_test.go`, `gapi/ai_config_contract_test.go`: update contract and config tests.
- `web/frontend/src/admin/types.ts`: add prompt template types.
- `web/frontend/src/admin/ai/polish.ts`: add candidate operation helpers and raw fallback-friendly types.
- `web/frontend/src/admin/ai/polish.test.ts`: test helpers.
- `web/frontend/src/views/admin/AdminAISettingsView.vue`: add advanced prompt template editor and reset controls.
- `web/frontend/src/views/admin/AdminArticleEditorView.vue`: add the right-side AI candidate drawer and selection-safe apply flow.
- `web/frontend/src/assets/content.css`: add drawer layout styles if existing utility classes are not enough.
- `web/frontend/src/admin/adminUiPolish.test.ts`: extend source-level UI contracts for settings and drawer.

## Task 1: Add SDK Dependencies

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

- [ ] **Step 1: Add official SDK modules**

Run:

```bash
go get github.com/openai/openai-go/v3@v3.39.0 github.com/anthropics/anthropic-sdk-go@v1.50.1
```

Expected: `go.mod` contains both SDKs and `go.sum` contains their checksums.

- [ ] **Step 2: Verify dependency graph resolves**

Run:

```bash
go mod tidy
go test ./internal/ai -run '^$' -count=1
```

Expected: command exits 0. `go test ./internal/ai -run '^$'` may report `[no tests to run]`, but the package must compile.

- [ ] **Step 3: Commit dependency update**

Run:

```bash
git add go.mod go.sum
git commit -m "chore: add ai provider sdks"
```

## Task 2: Persist Prompt Templates In AI Config

**Files:**
- Create: `db/migration/000013_add_ai_prompt_templates.up.sql`
- Create: `db/migration/000013_add_ai_prompt_templates.down.sql`
- Modify: `db/query/ai_provider_config.sql`
- Generate: `db/sqlc/ai_provider_config.sql.go`
- Generate: `db/sqlc/models.go`
- Generate: `db/mock/store.go`
- Modify: `proto/rpc_polish_text.proto`
- Generate: `pb/*.go`
- Modify: `gapi/ai_config_contract_test.go`
- Modify: `gapi/rpc_polish_text_test.go`

- [ ] **Step 1: Write failing contract tests**

Update `gapi/ai_config_contract_test.go` to require the new migration, SQL, and proto fields:

```go
func TestAIConfigPromptTemplateContracts(t *testing.T) {
	protoSource, err := os.ReadFile("../proto/rpc_polish_text.proto")
	require.NoError(t, err)
	querySource, err := os.ReadFile("../db/query/ai_provider_config.sql")
	require.NoError(t, err)
	migrationSource, err := os.ReadFile("../db/migration/000013_add_ai_prompt_templates.up.sql")
	require.NoError(t, err)

	require.Contains(t, string(protoSource), "map<string, string> prompt_templates =")
	require.Contains(t, string(querySource), "prompt_templates")
	require.Contains(t, string(migrationSource), "ADD COLUMN prompt_templates jsonb")
}
```

- [ ] **Step 2: Run contract test and confirm it fails**

Run:

```bash
go test ./gapi -run TestAIConfigPromptTemplateContracts -count=1
```

Expected: FAIL because the migration/proto/query fields do not exist yet.

- [ ] **Step 3: Add migration files**

Create `db/migration/000013_add_ai_prompt_templates.up.sql`:

```sql
ALTER TABLE ai_provider_configs
    ADD COLUMN prompt_templates jsonb NOT NULL DEFAULT '{}'::jsonb;
```

Create `db/migration/000013_add_ai_prompt_templates.down.sql`:

```sql
ALTER TABLE ai_provider_configs
    DROP COLUMN IF EXISTS prompt_templates;
```

- [ ] **Step 4: Update SQL query**

Modify `db/query/ai_provider_config.sql` so `UpsertAIProviderConfig` inserts and updates `prompt_templates`:

```sql
INSERT INTO ai_provider_configs (
    purpose,
    provider,
    api_protocol,
    base_url,
    model,
    api_key_ciphertext,
    timeout_ms,
    max_input_chars,
    max_context_chars,
    max_suggestions,
    prompt_templates,
    enabled,
    updated_by
) VALUES (
    sqlc.arg(purpose),
    sqlc.arg(provider),
    sqlc.arg(api_protocol),
    sqlc.arg(base_url),
    sqlc.arg(model),
    sqlc.arg(api_key_ciphertext),
    sqlc.arg(timeout_ms),
    sqlc.arg(max_input_chars),
    sqlc.arg(max_context_chars),
    sqlc.arg(max_suggestions),
    sqlc.arg(prompt_templates),
    sqlc.arg(enabled),
    sqlc.narg(updated_by)
)
ON CONFLICT (purpose) DO UPDATE
SET provider = EXCLUDED.provider,
    api_protocol = EXCLUDED.api_protocol,
    base_url = EXCLUDED.base_url,
    model = EXCLUDED.model,
    api_key_ciphertext = EXCLUDED.api_key_ciphertext,
    timeout_ms = EXCLUDED.timeout_ms,
    max_input_chars = EXCLUDED.max_input_chars,
    max_context_chars = EXCLUDED.max_context_chars,
    max_suggestions = EXCLUDED.max_suggestions,
    prompt_templates = EXCLUDED.prompt_templates,
    enabled = EXCLUDED.enabled,
    updated_by = EXCLUDED.updated_by,
    updated_at = now()
RETURNING *;
```

- [ ] **Step 5: Add proto fields**

Modify `proto/rpc_polish_text.proto`:

```proto
message GetAIConfigResponse {
  string provider = 1;
  string base_url = 2;
  string model = 3;
  bool api_key_configured = 4;
  bool enabled = 5;
  string timeout = 6;
  int32 max_input_chars = 7;
  int32 max_context_chars = 8;
  int32 max_suggestions = 9;
  string source = 10;
  string api_protocol = 11;
  map<string, string> prompt_templates = 12;
  map<string, string> default_prompt_templates = 13;
}

message UpdateAIConfigRequest {
  string provider = 1;
  string base_url = 2;
  string model = 3;
  string api_key = 4;
  string timeout = 5;
  int32 max_input_chars = 6;
  int32 max_context_chars = 7;
  int32 max_suggestions = 8;
  bool enabled = 9;
  bool clear_api_key = 10;
  string api_protocol = 11;
  map<string, string> prompt_templates = 12;
}
```

- [ ] **Step 6: Regenerate sqlc, mocks, and proto**

Run:

```bash
make sqlc
make mock
make proto
```

Expected: generated files compile with the new `PromptTemplates` fields.

- [ ] **Step 7: Verify contract test passes**

Run:

```bash
go test ./gapi -run 'TestAIConfigPromptTemplateContracts|TestAIConfigRuntimeMutationContracts' -count=1
```

Expected: PASS.

- [ ] **Step 8: Commit persistence contract**

Run:

```bash
git add db/migration/000013_add_ai_prompt_templates.* db/query/ai_provider_config.sql db/sqlc db/mock/store.go proto pb gapi/ai_config_contract_test.go
git commit -m "feat: persist ai prompt templates"
```

## Task 3: Build Prompt Template And Parser Core

**Files:**
- Modify: `internal/ai/types.go`
- Create: `internal/ai/templates.go`
- Create: `internal/ai/templates_test.go`
- Create: `internal/ai/parser.go`
- Create: `internal/ai/parser_test.go`

- [ ] **Step 1: Write failing template tests**

Create `internal/ai/templates_test.go`:

```go
package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizePromptTemplatesFillsDefaults(t *testing.T) {
	templates := NormalizePromptTemplates(map[string]string{
		ModeImprove: "custom improve {{text}}",
	})

	require.Equal(t, "custom improve {{text}}", templates[ModeImprove])
	require.NotEmpty(t, templates[ModeShorten])
	require.NotEmpty(t, templates[ModeExpand])
	require.NotEmpty(t, templates[ModeTitleCandidates])
	require.NotEmpty(t, templates[ModeSummaryCandidates])
}

func TestRenderPromptTemplateReplacesVariables(t *testing.T) {
	rendered := RenderPromptTemplate("{{mode}} {{text}} {{article_title}} {{max_suggestions}}", PromptRenderData{
		Mode:           ModeImprove,
		Text:           "原文",
		ArticleTitle:   "标题",
		MaxSuggestions: 3,
	})

	require.Equal(t, "improve 原文 标题 3", rendered)
}

func TestDefaultPromptTemplatesAskForSuggestionsJSON(t *testing.T) {
	for mode, template := range DefaultPromptTemplates() {
		require.Contains(t, template, `"suggestions"`, mode)
		require.Contains(t, strings.ToLower(template), "json", mode)
	}
}
```

- [ ] **Step 2: Write failing parser tests**

Create `internal/ai/parser_test.go`:

```go
package ai

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSuggestionsReadsJSONEnvelope(t *testing.T) {
	suggestions, err := ParseSuggestions(`{"suggestions":[{"content":"A","reason":"R"},{"content":"B"}]}`, 1)

	require.NoError(t, err)
	require.Equal(t, []Suggestion{{Content: "A", Reason: "R"}}, suggestions)
}

func TestParseSuggestionsFallsBackToRawText(t *testing.T) {
	suggestions, err := ParseSuggestions("普通文本候选", 3)

	require.NoError(t, err)
	require.Equal(t, []Suggestion{{Content: "普通文本候选"}}, suggestions)
}

func TestParseSuggestionsFallsBackWhenJSONHasNoUsableSuggestions(t *testing.T) {
	suggestions, err := ParseSuggestions(`{"suggestions":[{"content":"   "}],"note":"raw"}`, 3)

	require.NoError(t, err)
	require.Equal(t, []Suggestion{{Content: `{"suggestions":[{"content":"   "}],"note":"raw"}`}}, suggestions)
}

func TestParseSuggestionsRejectsEmptyOutput(t *testing.T) {
	_, err := ParseSuggestions("  ", 3)

	require.ErrorIs(t, err, ErrMalformedResponse)
}
```

- [ ] **Step 3: Run tests and confirm they fail**

Run:

```bash
go test ./internal/ai -run 'TestNormalizePromptTemplates|TestRenderPromptTemplate|TestDefaultPromptTemplates|TestParseSuggestions' -count=1
```

Expected: FAIL because the template and parser functions are not implemented.

- [ ] **Step 4: Implement template helpers**

Add `PromptRenderData` and template helpers to `internal/ai/templates.go`:

```go
package ai

import (
	"sort"
	"strconv"
	"strings"
)

type PromptRenderData struct {
	Mode           string
	Target         string
	Text           string
	ArticleTitle   string
	ArticleSummary string
	ArticleExcerpt string
	Locale         string
	MaxSuggestions int
}

func DefaultPromptTemplates() map[string]string {
	return map[string]string{
		ModeImprove: `请润色以下内容。返回 JSON：{"suggestions":[{"content":"...","reason":"..."}]}。
mode={{mode}}
target={{target}}
locale={{locale}}
article_title={{article_title}}
article_summary={{article_summary}}
text:
{{text}}`,
		ModeShorten: `请精简以下内容。返回 JSON：{"suggestions":[{"content":"...","reason":"..."}]}。
locale={{locale}}
text:
{{text}}`,
		ModeExpand: `请扩写以下内容。返回 JSON：{"suggestions":[{"content":"...","reason":"..."}]}。
locale={{locale}}
text:
{{text}}`,
		ModeTitleCandidates: `请基于文章上下文生成标题候选。最多 {{max_suggestions}} 个。返回 JSON：{"suggestions":[{"content":"...","reason":"..."}]}。
article_title={{article_title}}
article_summary={{article_summary}}
article_excerpt:
{{article_excerpt}}`,
		ModeSummaryCandidates: `请基于文章上下文生成摘要候选。最多 {{max_suggestions}} 个。返回 JSON：{"suggestions":[{"content":"...","reason":"..."}]}。
article_title={{article_title}}
article_summary={{article_summary}}
article_excerpt:
{{article_excerpt}}`,
	}
}

func PromptTemplateKeys() []string {
	keys := []string{ModeImprove, ModeShorten, ModeExpand, ModeTitleCandidates, ModeSummaryCandidates}
	sort.Strings(keys)
	return keys
}

func NormalizePromptTemplates(values map[string]string) map[string]string {
	defaults := DefaultPromptTemplates()
	normalized := make(map[string]string, len(defaults))
	for key, value := range defaults {
		normalized[key] = value
		if custom := strings.TrimSpace(values[key]); custom != "" {
			normalized[key] = custom
		}
	}
	return normalized
}

func RenderPromptTemplate(template string, data PromptRenderData) string {
	replacements := map[string]string{
		"{{mode}}":            data.Mode,
		"{{target}}":          data.Target,
		"{{text}}":            data.Text,
		"{{article_title}}":   data.ArticleTitle,
		"{{article_summary}}": data.ArticleSummary,
		"{{article_excerpt}}": data.ArticleExcerpt,
		"{{locale}}":          data.Locale,
		"{{max_suggestions}}": strconv.Itoa(data.MaxSuggestions),
	}
	rendered := template
	for token, value := range replacements {
		rendered = strings.ReplaceAll(rendered, token, value)
	}
	return rendered
}
```

- [ ] **Step 5: Implement parser**

Add `internal/ai/parser.go`:

```go
package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

type suggestionEnvelope struct {
	Suggestions []Suggestion `json:"suggestions"`
}

func ParseSuggestions(raw string, max int) ([]Suggestion, error) {
	content := strings.TrimSpace(raw)
	if content == "" {
		return nil, fmt.Errorf("%w: empty provider content", ErrMalformedResponse)
	}

	var envelope suggestionEnvelope
	if err := json.Unmarshal([]byte(content), &envelope); err == nil {
		suggestions := normalizeSuggestions(envelope.Suggestions, max)
		if len(suggestions) > 0 {
			return suggestions, nil
		}
	}

	return []Suggestion{{Content: content}}, nil
}
```

Remove the duplicate `suggestionEnvelope` type from `internal/ai/openai_compatible.go` when this is implemented.

- [ ] **Step 6: Verify prompt and parser tests pass**

Run:

```bash
gofmt -w internal/ai
go test ./internal/ai -run 'TestNormalizePromptTemplates|TestRenderPromptTemplate|TestDefaultPromptTemplates|TestParseSuggestions' -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit prompt and parser core**

Run:

```bash
git add internal/ai
git commit -m "feat: add ai prompt template parsing"
```

## Task 4: Refactor AI Service And Official SDK Adapters

**Files:**
- Modify: `internal/ai/types.go`
- Create: `internal/ai/service.go`
- Create: `internal/ai/service_test.go`
- Create: `internal/ai/openai_adapter.go`
- Create: `internal/ai/openai_adapter_test.go`
- Create: `internal/ai/anthropic_adapter.go`
- Create: `internal/ai/anthropic_adapter_test.go`
- Delete or leave unused after replacement: `internal/ai/openai_compatible.go`
- Modify: `internal/ai/openai_compatible_test.go`

- [ ] **Step 1: Write failing service tests**

Create `internal/ai/service_test.go`:

```go
package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeProviderAdapter struct {
	request GenerateRequest
	output  string
	err     error
}

func (adapter *fakeProviderAdapter) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	adapter.request = req
	if adapter.err != nil {
		return GenerateResponse{}, adapter.err
	}
	return GenerateResponse{Content: adapter.output, Model: req.Model}, nil
}

func (adapter *fakeProviderAdapter) ListModels(ctx context.Context) ([]Model, error) {
	return []Model{{ID: "writer-model"}}, nil
}

func TestPolishServiceRendersTemplateAndParsesJSON(t *testing.T) {
	adapter := &fakeProviderAdapter{output: `{"suggestions":[{"content":"更好","reason":"更顺"}]}`}
	service := NewPolishService(ServiceConfig{
		Provider:        "openai",
		APIProtocol:     APIProtocolChatCompletions,
		Model:           "writer-model",
		MaxInputChars:   6000,
		MaxContextChars: 4000,
		MaxSuggestions:  3,
		PromptTemplates: map[string]string{ModeImprove: "mode={{mode}} text={{text}}"},
	}, func(ServiceConfig) (ProviderAdapter, error) {
		return adapter, nil
	})

	resp, err := service.Polish(context.Background(), PolishRequest{
		Mode:   ModeImprove,
		Target: TargetContentSelection,
		Text:   "原文",
	})

	require.NoError(t, err)
	require.Equal(t, "mode=improve text=原文", adapter.request.Prompt)
	require.Equal(t, "更好", resp.Suggestions[0].Content)
	require.Equal(t, "writer-model", resp.Model)
}

func TestPolishServiceFallsBackToRawCandidate(t *testing.T) {
	adapter := &fakeProviderAdapter{output: "直接返回文本"}
	service := NewPolishService(ServiceConfig{
		Provider:       "openai",
		APIProtocol:    APIProtocolResponses,
		Model:          "writer-model",
		MaxInputChars:  6000,
		MaxSuggestions: 3,
	}, func(ServiceConfig) (ProviderAdapter, error) {
		return adapter, nil
	})

	resp, err := service.Polish(context.Background(), PolishRequest{
		Mode:   ModeImprove,
		Target: TargetContentSelection,
		Text:   "原文",
	})

	require.NoError(t, err)
	require.Equal(t, []Suggestion{{Content: "直接返回文本"}}, resp.Suggestions)
}
```

- [ ] **Step 2: Define service and adapter types**

Update `internal/ai/types.go` with:

```go
const (
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
)

type ServiceConfig struct {
	Provider         string
	APIProtocol      string
	BaseURL          string
	APIKey           string
	Model            string
	Timeout          time.Duration
	MaxInputChars    int
	MaxContextChars  int
	MaxSuggestions   int
	PromptTemplates  map[string]string
	HTTPProxyAddress string
}

type GenerateRequest struct {
	Protocol string
	Model    string
	Prompt   string
}

type GenerateResponse struct {
	Content string
	Model   string
}

type ProviderAdapter interface {
	Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
	ListModels(ctx context.Context) ([]Model, error)
}

type ProviderFactory func(ServiceConfig) (ProviderAdapter, error)
```

Keep `TextPolisher` so `gapi.Server` tests can still inject fakes.

- [ ] **Step 3: Implement `PolishService`**

Create `internal/ai/service.go`:

```go
package ai

import (
	"context"
	"fmt"
)

type PolishService struct {
	config  ServiceConfig
	factory ProviderFactory
}

func NewPolishService(config ServiceConfig, factory ProviderFactory) TextPolisher {
	if factory == nil {
		factory = NewProviderAdapter
	}
	config.PromptTemplates = NormalizePromptTemplates(config.PromptTemplates)
	return &PolishService{config: config, factory: factory}
}

func (service *PolishService) Polish(ctx context.Context, req PolishRequest) (PolishResponse, error) {
	req = req.normalized()
	if service.disabled() {
		return PolishResponse{}, ErrDisabled
	}
	if err := validateRequest(req, service.config.MaxInputChars); err != nil {
		return PolishResponse{}, err
	}

	req.ArticleTitle = limitRunes(req.ArticleTitle, service.config.MaxContextChars)
	req.ArticleSummary = limitRunes(req.ArticleSummary, service.config.MaxContextChars)
	req.ArticleExcerpt = limitRunes(req.ArticleExcerpt, service.config.MaxContextChars)

	adapter, err := service.factory(service.config)
	if err != nil {
		return PolishResponse{}, err
	}

	prompt := RenderPromptTemplate(service.config.PromptTemplates[req.Mode], PromptRenderData{
		Mode:           req.Mode,
		Target:         req.Target,
		Text:           req.Text,
		ArticleTitle:   req.ArticleTitle,
		ArticleSummary: req.ArticleSummary,
		ArticleExcerpt: req.ArticleExcerpt,
		Locale:         req.Locale,
		MaxSuggestions: service.config.MaxSuggestions,
	})

	generated, err := adapter.Generate(ctx, GenerateRequest{
		Protocol: service.config.APIProtocol,
		Model:    service.config.Model,
		Prompt:   prompt,
	})
	if err != nil {
		return PolishResponse{}, err
	}

	suggestions, err := ParseSuggestions(generated.Content, service.config.MaxSuggestions)
	if err != nil {
		return PolishResponse{}, err
	}

	return PolishResponse{
		Suggestions: suggestions,
		Mode:        req.Mode,
		Target:      req.Target,
		Model:       generated.Model,
	}, nil
}

func (service *PolishService) disabled() bool {
	return service.config.BaseURL == "" || service.config.APIKey == "" || service.config.Model == ""
}

func NewProviderAdapter(config ServiceConfig) (ProviderAdapter, error) {
	switch normalizeProvider(config.Provider) {
	case ProviderAnthropic:
		return NewAnthropicAdapter(config), nil
	case ProviderOpenAI:
		return NewOpenAIAdapter(config), nil
	default:
		return nil, fmt.Errorf("%w: unsupported provider", ErrInvalidInput)
	}
}
```

- [ ] **Step 4: Implement provider normalization**

Add this helper in `internal/ai/types.go` or `internal/ai/service.go`:

```go
func normalizeProvider(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "anthropic", "claude":
		return ProviderAnthropic
	default:
		return ProviderOpenAI
	}
}
```

- [ ] **Step 5: Write OpenAI adapter request-shape tests**

Create `internal/ai/openai_adapter_test.go` using `httptest.Server`. Test Chat Completions hits `/chat/completions` and Responses hits `/responses`:

```go
func TestOpenAIAdapterUsesChatCompletions(t *testing.T) {
	var gotPath string
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		require.Equal(t, "Bearer secret-key", r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"raw candidate"}}]}`))
	}))
	defer provider.Close()

	adapter := NewOpenAIAdapter(ServiceConfig{
		BaseURL:     provider.URL,
		APIKey:      "secret-key",
		APIProtocol: APIProtocolChatCompletions,
		Model:       "writer-model",
	})

	resp, err := adapter.Generate(context.Background(), GenerateRequest{
		Protocol: APIProtocolChatCompletions,
		Model:    "writer-model",
		Prompt:   "hello",
	})

	require.NoError(t, err)
	require.Equal(t, "/chat/completions", gotPath)
	require.Equal(t, "raw candidate", resp.Content)
}
```

- [ ] **Step 6: Write Anthropic adapter request-shape test**

Create `internal/ai/anthropic_adapter_test.go`:

```go
func TestAnthropicAdapterUsesMessages(t *testing.T) {
	var gotPath string
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		require.Equal(t, "secret-key", r.Header.Get("x-api-key"))
		_, _ = w.Write([]byte(`{"id":"msg_1","type":"message","role":"assistant","model":"claude-test","content":[{"type":"text","text":"raw candidate"}],"stop_reason":"end_turn","stop_sequence":null,"usage":{"input_tokens":1,"output_tokens":1}}`))
	}))
	defer provider.Close()

	adapter := NewAnthropicAdapter(ServiceConfig{
		BaseURL: provider.URL,
		APIKey:  "secret-key",
		Model:   "claude-test",
	})

	resp, err := adapter.Generate(context.Background(), GenerateRequest{
		Protocol: APIProtocolMessages,
		Model:    "claude-test",
		Prompt:   "hello",
	})

	require.NoError(t, err)
	require.Equal(t, "/v1/messages", gotPath)
	require.Equal(t, "raw candidate", resp.Content)
}
```

- [ ] **Step 7: Implement `OpenAIAdapter`**

Create `internal/ai/openai_adapter.go`. Use the official SDK:

```go
package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

type OpenAIAdapter struct {
	config ServiceConfig
	client openai.Client
}

func NewOpenAIAdapter(config ServiceConfig) ProviderAdapter {
	opts := []option.RequestOption{option.WithAPIKey(config.APIKey)}
	if strings.TrimSpace(config.BaseURL) != "" {
		opts = append(opts, option.WithBaseURL(strings.TrimRight(config.BaseURL, "/")))
	}
	return &OpenAIAdapter{config: config, client: openai.NewClient(opts...)}
}

func (adapter *OpenAIAdapter) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	switch req.Protocol {
	case APIProtocolResponses:
		resp, err := adapter.client.Responses.New(ctx, responses.ResponseNewParams{
			Model: responses.ResponseModel(req.Model),
			Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(req.Prompt)},
		})
		if err != nil {
			return GenerateResponse{}, fmt.Errorf("%w: openai responses request failed", ErrProviderFailure)
		}
		return GenerateResponse{Content: strings.TrimSpace(resp.OutputText()), Model: req.Model}, nil
	default:
		resp, err := adapter.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model: openai.ChatModel(req.Model),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(req.Prompt),
			},
		})
		if err != nil {
			return GenerateResponse{}, fmt.Errorf("%w: openai chat request failed", ErrProviderFailure)
		}
		if len(resp.Choices) == 0 {
			return GenerateResponse{}, fmt.Errorf("%w: missing openai choices", ErrMalformedResponse)
		}
		return GenerateResponse{Content: strings.TrimSpace(resp.Choices[0].Message.Content), Model: req.Model}, nil
	}
}

func (adapter *OpenAIAdapter) ListModels(ctx context.Context) ([]Model, error) {
	page, err := adapter.client.Models.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: openai models request failed", ErrProviderFailure)
	}
	models := make([]Model, 0, len(page.Data))
	for _, item := range page.Data {
		if id := strings.TrimSpace(item.ID); id != "" {
			models = append(models, Model{ID: id})
		}
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("%w: empty models list", ErrMalformedResponse)
	}
	return models, nil
}
```

Before committing, compile the adapter against the local module cache and align the concrete SDK type names with the installed versions while preserving this adapter boundary.

- [ ] **Step 8: Implement `AnthropicAdapter`**

Create `internal/ai/anthropic_adapter.go`:

```go
package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type AnthropicAdapter struct {
	config ServiceConfig
	client anthropic.Client
}

func NewAnthropicAdapter(config ServiceConfig) ProviderAdapter {
	opts := []option.RequestOption{option.WithAPIKey(config.APIKey)}
	if strings.TrimSpace(config.BaseURL) != "" {
		opts = append(opts, option.WithBaseURL(strings.TrimRight(config.BaseURL, "/")))
	}
	return &AnthropicAdapter{config: config, client: anthropic.NewClient(opts...)}
}

func (adapter *AnthropicAdapter) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	msg, err := adapter.client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 2048,
		Model:     anthropic.Model(req.Model),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(req.Prompt)),
		},
	})
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("%w: anthropic messages request failed", ErrProviderFailure)
	}

	var parts []string
	for _, block := range msg.Content {
		if text := strings.TrimSpace(block.Text); text != "" {
			parts = append(parts, text)
		}
	}
	if len(parts) == 0 {
		return GenerateResponse{}, fmt.Errorf("%w: missing anthropic content", ErrMalformedResponse)
	}
	return GenerateResponse{Content: strings.Join(parts, "\n"), Model: req.Model}, nil
}

func (adapter *AnthropicAdapter) ListModels(ctx context.Context) ([]Model, error) {
	return nil, fmt.Errorf("%w: anthropic model listing is unsupported", ErrProviderFailure)
}
```

- [ ] **Step 9: Remove old combined implementation**

Delete `internal/ai/openai_compatible.go` after the new tests cover Chat Completions, Responses, Anthropic Messages, model listing, and raw fallback. Replace references to `NewOpenAICompatiblePolisher` with `NewPolishService`.

- [ ] **Step 10: Verify AI package**

Run:

```bash
gofmt -w internal/ai
go test ./internal/ai -count=1
```

Expected: PASS.

- [ ] **Step 11: Commit service refactor**

Run:

```bash
git add internal/ai go.mod go.sum
git commit -m "feat: refactor ai polish provider adapters"
```

## Task 5: Wire Prompt Templates Through gAPI

**Files:**
- Modify: `gapi/ai_config.go`
- Modify: `gapi/rpc_get_ai_config.go`
- Modify: `gapi/rpc_update_ai_config.go`
- Modify: `gapi/rpc_polish_text.go`
- Modify: `gapi/rpc_list_ai_models.go`
- Modify: `gapi/rpc_polish_text_test.go`

- [ ] **Step 1: Write failing config round-trip test**

Extend `TestUpdateAIConfigStoresEncryptedSecretAndMasksResponse` or add:

```go
func TestUpdateAIConfigStoresPromptTemplates(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	store.EXPECT().
		GetAIProviderConfig(gomock.Any(), "ai_polish").
		Return(db.AiProviderConfig{}, pgx.ErrNoRows)
	store.EXPECT().
		UpsertAIProviderConfig(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, arg db.UpsertAIProviderConfigParams) (db.AiProviderConfig, error) {
			require.JSONEq(t, `{"improve":"custom {{text}}"}`, string(arg.PromptTemplates))
			return db.AiProviderConfig{
				Purpose:          arg.Purpose,
				Provider:         arg.Provider,
				ApiProtocol:      arg.ApiProtocol,
				BaseUrl:          arg.BaseUrl,
				Model:            arg.Model,
				ApiKeyCiphertext: arg.ApiKeyCiphertext,
				TimeoutMs:        arg.TimeoutMs,
				MaxInputChars:    arg.MaxInputChars,
				MaxContextChars:  arg.MaxContextChars,
				MaxSuggestions:   arg.MaxSuggestions,
				PromptTemplates:  arg.PromptTemplates,
				Enabled:          arg.Enabled,
				UpdatedBy:        arg.UpdatedBy,
			}, nil
		})

	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)
	resp, err := server.UpdateAIConfig(ctx, &pb.UpdateAIConfigRequest{
		Provider:        "openai",
		ApiProtocol:     ai.APIProtocolChatCompletions,
		BaseUrl:         "https://ai.example.com/v1",
		Model:           "writer-model",
		ApiKey:          "new-secret",
		Timeout:         "30s",
		MaxInputChars:   6000,
		MaxContextChars: 4000,
		MaxSuggestions:  3,
		Enabled:         true,
		PromptTemplates: map[string]string{ai.ModeImprove: "custom {{text}}"},
	})

	require.NoError(t, err)
	require.Equal(t, "custom {{text}}", resp.GetPromptTemplates()[ai.ModeImprove])
	require.NotEmpty(t, resp.GetDefaultPromptTemplates()[ai.ModeImprove])
}
```

- [ ] **Step 2: Run gapi prompt tests and confirm failure**

Run:

```bash
go test ./gapi -run 'TestUpdateAIConfigStoresPromptTemplates|TestPolishTextUsesDatabaseConfig' -count=1
```

Expected: FAIL until config structs and persistence support prompt templates.

- [ ] **Step 3: Extend `resolvedAIConfig`**

Add to `gapi/ai_config.go`:

```go
PromptTemplates map[string]string
```

Set runtime defaults:

```go
PromptTemplates: ai.DefaultPromptTemplates(),
```

- [ ] **Step 4: Decode and encode prompt template JSON**

Add helpers in `gapi/ai_config.go`:

```go
func decodePromptTemplates(raw []byte) map[string]string {
	if len(raw) == 0 {
		return ai.DefaultPromptTemplates()
	}
	var values map[string]string
	if err := json.Unmarshal(raw, &values); err != nil {
		return ai.DefaultPromptTemplates()
	}
	return ai.NormalizePromptTemplates(values)
}

func encodePromptTemplates(values map[string]string) []byte {
	encoded, err := json.Marshal(ai.NormalizePromptTemplates(values))
	if err != nil {
		return []byte(`{}`)
	}
	return encoded
}
```

- [ ] **Step 5: Include templates in response and save params**

Update `toResponse`:

```go
PromptTemplates:        ai.NormalizePromptTemplates(cfg.PromptTemplates),
DefaultPromptTemplates: ai.DefaultPromptTemplates(),
```

Update `saveAIConfig`:

```go
PromptTemplates: encodePromptTemplates(cfg.PromptTemplates),
```

Update `aiConfigFromRow`:

```go
PromptTemplates: decodePromptTemplates(row.PromptTemplates),
```

- [ ] **Step 6: Build service config in `PolishText`**

Replace `ai.NewOpenAICompatiblePolisher(...)` in `gapi/rpc_polish_text.go` with:

```go
polisher = ai.NewPolishService(ai.ServiceConfig{
	Provider:         cfg.Provider,
	APIProtocol:      cfg.APIProtocol,
	BaseURL:          cfg.BaseURL,
	APIKey:           cfg.APIKey,
	Model:            cfg.Model,
	Timeout:          cfg.Timeout,
	MaxInputChars:    cfg.MaxInputChars,
	MaxContextChars:  cfg.MaxContextChars,
	MaxSuggestions:   cfg.MaxSuggestions,
	PromptTemplates:  cfg.PromptTemplates,
	HTTPProxyAddress: server.config.HTTPProxyAddr,
}, nil)
```

- [ ] **Step 7: Update model listing**

In `gapi/rpc_list_ai_models.go`, construct adapter through `ai.NewProviderAdapter`. For Anthropic, return:

```go
status.Error(codes.Unimplemented, "Anthropic model listing is not supported yet")
```

Keep OpenAI model listing behavior.

- [ ] **Step 8: Verify gapi tests**

Run:

```bash
gofmt -w gapi
go test ./gapi -run 'TestPolishText|TestGetAIConfig|TestUpdateAIConfig|TestListAIModels|TestAIConfig' -count=1
```

Expected: PASS.

- [ ] **Step 9: Commit gAPI wiring**

Run:

```bash
git add gapi
git commit -m "feat: wire ai prompt templates through api"
```

## Task 6: Update Frontend Types And AI Settings Prompt Editor

**Files:**
- Modify: `web/frontend/src/admin/types.ts`
- Modify: `web/frontend/src/views/admin/AdminAISettingsView.vue`
- Modify: `web/frontend/src/admin/adminUiPolish.test.ts`

- [ ] **Step 1: Write failing UI contract test**

Extend the existing `admin routes include a dedicated AI settings page` test in `web/frontend/src/admin/adminUiPolish.test.ts`:

```ts
expect(pageSource).toContain('promptTemplates')
expect(pageSource).toContain('defaultPromptTemplates')
expect(pageSource).toContain('高级提示词')
expect(pageSource).toContain('resetPromptTemplate')
expect(pageSource).toContain('resetAllPromptTemplates')
expect(typesSource).toContain('prompt_templates')
expect(typesSource).toContain('default_prompt_templates')
```

- [ ] **Step 2: Run frontend contract test and confirm failure**

Run:

```bash
cd web/frontend && bun test src/admin/adminUiPolish.test.ts
```

Expected: FAIL because the settings page and types do not contain prompt template fields.

- [ ] **Step 3: Extend frontend types**

Modify `web/frontend/src/admin/types.ts`:

```ts
export type AdminAIPromptTemplates = Partial<Record<AdminAIPolishMode, string>>

export interface AdminAIConfigResponse {
  provider: string
  api_protocol: AdminAIProtocol
  base_url: string
  model: string
  api_key_configured: boolean
  enabled: boolean
  timeout: string
  max_input_chars: number
  max_context_chars: number
  max_suggestions: number
  source: 'runtime_env' | string
  prompt_templates: AdminAIPromptTemplates
  default_prompt_templates: AdminAIPromptTemplates
}

export interface AdminAIConfigUpdateRequest {
  provider: string
  api_protocol: AdminAIProtocol
  base_url: string
  model: string
  api_key?: string
  timeout: string
  max_input_chars: number
  max_context_chars: number
  max_suggestions: number
  enabled: boolean
  clear_api_key?: boolean
  prompt_templates?: AdminAIPromptTemplates
}
```

- [ ] **Step 4: Add prompt template state**

In `AdminAISettingsView.vue`, add:

```ts
const promptModes: AdminAIPolishMode[] = [
  'improve',
  'shorten',
  'expand',
  'title_candidates',
  'summary_candidates'
]
const promptTemplates = ref<Record<AdminAIPolishMode, string>>({
  improve: '',
  shorten: '',
  expand: '',
  title_candidates: '',
  summary_candidates: ''
})
const defaultPromptTemplates = ref<Record<AdminAIPolishMode, string>>({
  improve: '',
  shorten: '',
  expand: '',
  title_candidates: '',
  summary_candidates: ''
})
```

- [ ] **Step 5: Sync prompt templates from config**

Extend `syncForm`:

```ts
const nextDefaults = nextConfig?.default_prompt_templates ?? {}
const nextTemplates = nextConfig?.prompt_templates ?? {}
for (const mode of promptModes) {
  defaultPromptTemplates.value[mode] = nextDefaults[mode] || ''
  promptTemplates.value[mode] = nextTemplates[mode] || nextDefaults[mode] || ''
}
```

- [ ] **Step 6: Send prompt templates when saving**

Extend `updateAdminAIConfig` payload:

```ts
prompt_templates: { ...promptTemplates.value }
```

- [ ] **Step 7: Add reset helpers**

Add:

```ts
const resetPromptTemplate = (mode: AdminAIPolishMode) => {
  promptTemplates.value[mode] = defaultPromptTemplates.value[mode] || ''
}

const resetAllPromptTemplates = () => {
  for (const mode of promptModes) {
    resetPromptTemplate(mode)
  }
}
```

- [ ] **Step 8: Add advanced prompt UI**

Add a section below provider limits:

```vue
<section class="mt-5 rounded-archive border border-border bg-surface-raised p-4">
  <div class="flex flex-wrap items-center justify-between gap-3">
    <div>
      <h2 class="m-0 text-base font-black text-foreground">高级提示词</h2>
      <p class="m-0 mt-1 text-sm text-muted-foreground">站长完全控制每个操作的 prompt，可随时恢复内置模板。</p>
    </div>
    <AppButton type="button" variant="secondary" size="sm" @click="resetAllPromptTemplates">
      恢复全部默认
    </AppButton>
  </div>

  <div class="mt-4 grid gap-4">
    <label v-for="mode in promptModes" :key="mode" class="block space-y-2">
      <span class="flex items-center justify-between gap-3">
        <span class="text-sm font-bold text-foreground">{{ getAIPolishModeLabel(mode) }}</span>
        <AppButton type="button" variant="ghost" size="sm" @click="resetPromptTemplate(mode)">
          恢复默认
        </AppButton>
      </span>
      <textarea
        v-model="promptTemplates[mode]"
        rows="7"
        class="w-full resize-y rounded-archive border border-border bg-surface px-4 py-3 font-mono text-xs leading-6 text-foreground outline-none transition-colors focus:border-accent focus:ring-2 focus:ring-accent/20"
      />
    </label>
  </div>
</section>
```

Import `getAIPolishModeLabel` from `@/admin/ai/polish`.

- [ ] **Step 9: Verify frontend settings tests**

Run:

```bash
cd web/frontend && bun test src/admin/adminUiPolish.test.ts
cd web/frontend && bun run type-check
```

Expected: PASS.

- [ ] **Step 10: Commit settings UI**

Run:

```bash
git add web/frontend/src/admin/types.ts web/frontend/src/views/admin/AdminAISettingsView.vue web/frontend/src/admin/adminUiPolish.test.ts
git commit -m "feat: add ai prompt template settings"
```

## Task 7: Add Editor AI Candidate Drawer

**Files:**
- Modify: `web/frontend/src/admin/ai/polish.ts`
- Modify: `web/frontend/src/admin/ai/polish.test.ts`
- Modify: `web/frontend/src/views/admin/AdminArticleEditorView.vue`
- Modify: `web/frontend/src/assets/content.css`
- Modify: `web/frontend/src/admin/adminUiPolish.test.ts`

- [ ] **Step 1: Add failing helper tests**

Extend `web/frontend/src/admin/ai/polish.test.ts`:

```ts
import { createAIPolishSession, getAIPolishApplyLabel } from './polish'

test('creates pending content polish sessions', () => {
  expect(
    createAIPolishSession({
      mode: 'improve',
      target: 'content_selection',
      sourceText: '原文'
    })
  ).toMatchObject({
    mode: 'improve',
    target: 'content_selection',
    sourceText: '原文',
    status: 'loading',
    selectedSuggestionIndex: -1
  })
})

test('labels explicit apply actions', () => {
  expect(getAIPolishApplyLabel('content_selection')).toBe('替换选区')
  expect(getAIPolishApplyLabel('title')).toBe('应用到标题')
  expect(getAIPolishApplyLabel('summary')).toBe('应用到摘要')
})
```

- [ ] **Step 2: Add failing UI contract test**

Extend `adminUiPolish.test.ts`:

```ts
expect(source).toContain('aiDrawerOpen')
expect(source).toContain('aiPolishSession')
expect(source).toContain('admin-ai-drawer')
expect(source).toContain('replaceSelectionWithSuggestion')
expect(source).toContain('captureEditorSelection')
expect(source).toContain('restoreCapturedSelection')
expect(contentCss).toContain('.admin-ai-drawer')
```

- [ ] **Step 3: Run frontend tests and confirm failure**

Run:

```bash
cd web/frontend && bun test src/admin/ai/polish.test.ts src/admin/adminUiPolish.test.ts
```

Expected: FAIL because helpers and drawer are not implemented.

- [ ] **Step 4: Add helper types**

Modify `web/frontend/src/admin/ai/polish.ts`:

```ts
export type AIPolishSessionStatus = 'idle' | 'loading' | 'ready' | 'error'

export interface AIPolishSession {
  mode: AdminAIPolishMode
  target: AdminAIPolishTarget
  sourceText: string
  status: AIPolishSessionStatus
  selectedSuggestionIndex: number
}

export function createAIPolishSession(input: {
  mode: AdminAIPolishMode
  target: AdminAIPolishTarget
  sourceText: string
}): AIPolishSession {
  return {
    mode: input.mode,
    target: input.target,
    sourceText: input.sourceText,
    status: 'loading',
    selectedSuggestionIndex: -1
  }
}

export function getAIPolishApplyLabel(target: AdminAIPolishTarget) {
  if (target === 'title') return '应用到标题'
  if (target === 'summary') return '应用到摘要'
  return '替换选区'
}
```

- [ ] **Step 5: Capture CKEditor selection before async calls**

In `AdminArticleEditorView.vue`, add:

```ts
const capturedSelectionRange = ref<any | null>(null)

const captureEditorSelection = () => {
  const editor = editorInstance.value as any
  const firstRange = editor?.model.document.selection.getFirstRange()
  capturedSelectionRange.value = firstRange ? firstRange.clone() : null
}

const restoreCapturedSelection = () => {
  const editor = editorInstance.value as any
  const range = capturedSelectionRange.value
  if (!editor || !range) return false
  editor.model.change((writer: any) => {
    writer.setSelection(range)
  })
  return true
}
```

Call `captureEditorSelection()` inside `requestSelectedContentPolish` before `requestAIPolish`.

- [ ] **Step 6: Add drawer state**

Add:

```ts
const aiDrawerOpen = ref(false)
const aiPolishSession = ref<AIPolishSession | null>(null)
const selectedSuggestionIndex = ref(-1)
```

Set these in `requestAIPolish`:

```ts
aiDrawerOpen.value = true
aiPolishSession.value = createAIPolishSession({ mode, target, sourceText: normalizedText })
selectedSuggestionIndex.value = -1
```

When suggestions return:

```ts
if (aiPolishSession.value) aiPolishSession.value.status = 'ready'
```

In `catch`:

```ts
if (aiPolishSession.value) aiPolishSession.value.status = 'error'
```

- [ ] **Step 7: Replace selection explicitly**

Replace `applyContentSuggestion` with:

```ts
const replaceSelectionWithSuggestion = (suggestion: AdminAIPolishSuggestion) => {
  const editor = editorInstance.value as any
  if (!editor) return

  const restored = restoreCapturedSelection()
  editor.model.change((writer: any) => {
    if (!restored) {
      const position = editor.model.document.selection.getLastPosition()
      writer.setSelection(position)
    }
    editor.model.insertContent(writer.createText(suggestion.content), editor.model.document.selection)
  })
  editorData.value = editor.getData()
  toast.add({ severity: 'success', summary: '已替换选区', detail: '记得保存文章', life: 2200 })
}
```

- [ ] **Step 8: Make title and summary require explicit apply**

Use `selectedSuggestionIndex` for preview. Only call `applyFieldSuggestion` from the drawer's apply button:

```ts
const previewPolishSuggestion = (index: number) => {
  selectedSuggestionIndex.value = index
}
```

Do not write `articleTitle` or `articleSummary` in `previewPolishSuggestion`.

- [ ] **Step 9: Add drawer template**

Move the existing `polishPanelOpen` panel into an aside:

```vue
<aside
  v-if="aiDrawerOpen"
  class="admin-ai-drawer archive-surface rounded-archive p-4"
  aria-label="AI 候选"
>
  <div class="flex items-start justify-between gap-3">
    <div>
      <h2 class="m-0 text-base font-black text-foreground">{{ polishPanelTitle }}</h2>
      <p class="m-0 mt-1 text-xs font-semibold text-muted-foreground">
        {{ isPolishing ? '正在生成候选' : '选择候选，确认后才会改动文章' }}
      </p>
    </div>
    <AppButton variant="ghost" size="icon" aria-label="关闭 AI 候选" @click="closePolishPanel">
      <X class="size-4" aria-hidden="true" />
    </AppButton>
  </div>

  <div v-if="isPolishing" class="mt-4 text-sm font-semibold text-muted-foreground">
    正在准备候选...
  </div>

  <div v-else class="mt-4 space-y-3">
    <article
      v-for="(suggestion, index) in polishSuggestions"
      :key="`${suggestion.content}-${index}`"
      class="rounded-archive border border-border bg-surface p-3"
    >
      <p class="m-0 whitespace-pre-wrap text-sm leading-7 text-foreground">{{ suggestion.content }}</p>
      <p v-if="suggestion.reason" class="m-0 mt-2 text-xs font-semibold text-muted-foreground">
        {{ suggestion.reason }}
      </p>
      <div class="mt-3 flex flex-wrap gap-2">
        <AppButton size="sm" @click="applyPolishSuggestion(suggestion)">
          <Sparkles class="size-4" aria-hidden="true" />
          {{ getAIPolishApplyLabel(polishTarget) }}
        </AppButton>
        <AppButton variant="secondary" size="sm" @click="previewPolishSuggestion(index)">
          预览
        </AppButton>
        <AppButton variant="secondary" size="sm" @click="copyPolishSuggestion(suggestion)">
          <Copy class="size-4" aria-hidden="true" />
          复制
        </AppButton>
      </div>
    </article>
  </div>
</aside>
```

- [ ] **Step 10: Style drawer**

Add to `web/frontend/src/assets/content.css`:

```css
.admin-ai-drawer {
  max-height: var(--admin-editor-workbench-height);
  overflow-y: auto;
  overscroll-behavior: contain;
}

@container (min-width: 1192px) {
  .admin-editor-frame.has-ai-drawer {
    grid-template-columns:
      minmax(0, var(--admin-editor-shell-width))
      minmax(18rem, var(--admin-editor-settings-width));
  }
}
```

Render the AI drawer above the settings panel inside the existing right column on layouts where the editor and side panels share one secondary column.

- [ ] **Step 11: Verify frontend drawer**

Run:

```bash
cd web/frontend && bun test src/admin/ai/polish.test.ts src/admin/adminUiPolish.test.ts
cd web/frontend && bun run type-check
```

Expected: PASS.

- [ ] **Step 12: Commit editor drawer**

Run:

```bash
git add web/frontend/src/admin/ai web/frontend/src/views/admin/AdminArticleEditorView.vue web/frontend/src/assets/content.css web/frontend/src/admin/adminUiPolish.test.ts
git commit -m "feat: add ai candidate drawer"
```

## Task 8: Full Verification And Documentation

**Files:**
- Modify: `README.md`
- Modify: `.env.example`
- Modify: `docker-compose.yaml`
- Modify: `docker-compose.dev.yaml`
- Modify: `.github/workflows/test.yml`

- [ ] **Step 1: Update docs and example env**

Ensure docs mention:

```env
AI_POLISH_PROVIDER=openai
AI_POLISH_API_PROTOCOL=chat/completions
AI_POLISH_BASE_URL=https://api.openai.com/v1
AI_POLISH_API_KEY=
AI_POLISH_MODEL=
AI_POLISH_TIMEOUT=30s
AI_POLISH_MAX_INPUT_CHARS=6000
AI_POLISH_MAX_CONTEXT_CHARS=4000
AI_POLISH_MAX_SUGGESTIONS=3
```

Do not add real API keys.

- [ ] **Step 2: Run backend focused tests**

Run:

```bash
go test ./internal/ai ./gapi ./util -count=1
```

Expected: PASS.

- [ ] **Step 3: Run frontend focused tests**

Run:

```bash
cd web/frontend && bun test src/admin/ai/polish.test.ts src/admin/adminUiPolish.test.ts
cd web/frontend && bun run type-check
```

Expected: PASS.

- [ ] **Step 4: Run full frontend build**

Run:

```bash
cd web/frontend && bun run build
```

Expected: PASS.

- [ ] **Step 5: Run full backend tests if services are available**

Run:

```bash
make test
```

Expected: PASS when local PostgreSQL and Redis dependencies required by the suite are available. For missing local services, record the exact service error and rely on focused tests plus CI.

- [ ] **Step 6: Commit verification docs**

Run:

```bash
git add README.md .env.example docker-compose.yaml docker-compose.dev.yaml .github/workflows/test.yml
git commit -m "docs: document ai polish configuration"
```

- [ ] **Step 7: Final status check**

Run:

```bash
git status --short --branch
git log --oneline -6
```

Expected: working tree clean except ignored `.superpowers/`, and recent commits match the task commits above.
