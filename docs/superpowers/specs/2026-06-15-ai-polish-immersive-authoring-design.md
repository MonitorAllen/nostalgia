# AI Polish Immersive Authoring Design

## Context

The current AI polish implementation is present but not reliable enough for real writing. It already has admin-only endpoints, provider configuration, protocol choices, a CKEditor entry point, and a suggestion panel. The next iteration should treat the existing code as a foundation to refactor, not as a final shape.

The goal is an owner-only immersive writing assistant for the backend article editor. AI may generate candidates, but it must never overwrite article content, title, or summary without an explicit owner action.

## Decisions

- Use a single global AI configuration in the first phase.
- Use a `PolishService` that works in writing operations, not provider-specific protocol details.
- Implement provider protocols with official SDK adapters:
  - `OpenAIAdapter` uses `github.com/openai/openai-go/v3` for Chat Completions and Responses, including OpenAI-compatible providers through configurable Base URL.
  - `AnthropicAdapter` uses `github.com/anthropics/anthropic-sdk-go` for the native Messages API.
- Do not implement a hand-written HTTP adapter in the first phase. Add it later only if a provider cannot be reached through the main OpenAI or Anthropic protocols.
- Do not use third-party LLM orchestration frameworks. This feature is a focused writing tool, not an agent framework.
- Allow fully custom prompts. The owner controls prompt safety and quality.
- Ship built-in default templates so the feature works after configuration.
- Parse model output as JSON first. If JSON parsing fails, return the raw model text as one candidate instead of failing the whole polish request.
- Use a right-side AI candidate drawer on desktop. On narrow screens, degrade to a bottom drawer.

## Goals

- Make AI polish callable and debuggable through stable backend adapters.
- Support OpenAI Chat Completions, OpenAI Responses, and Anthropic Messages.
- Let the owner fully customize prompt templates per operation.
- Provide defaults for polish, shorten, expand, title candidates, and summary candidates.
- Preserve the editor content until the owner confirms a candidate.
- Keep title and summary candidates in preview or pending state until applied.
- Keep API keys server-side and masked in responses.
- Keep the implementation testable through small service, adapter, prompt, parser, and frontend helper units.

## Non-Goals

- No multi-provider preset manager in the first phase.
- No hand-written provider HTTP fallback in the first phase.
- No streaming output in the first phase.
- No autonomous article creation, saving, publishing, or deletion.
- No public-facing AI feature.
- No usage billing, quota dashboard, or cost analytics.
- No content safety enforcement beyond owner-controlled prompts.

## Backend Architecture

The backend should separate writing intent from provider transport:

```text
gapi.PolishText
  -> resolve AI config
  -> ai.PolishService
       -> PromptRenderer
       -> ProviderAdapter
            -> OpenAIAdapter
            -> AnthropicAdapter
       -> SuggestionParser
  -> normalized PolishTextResponse
```

`gapi` handles admin auth, config resolution, request validation, and gRPC error mapping. `internal/ai` owns prompt rendering, provider calls, output parsing, and typed errors.

The core provider interface should stay small:

```go
type ProviderAdapter interface {
    Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
    ListModels(ctx context.Context) ([]Model, error)
}
```

The polish service decides which adapter to use from `provider` and `api_protocol`. The rest of the application does not import SDK-specific types.

## Provider Protocols

`OpenAIAdapter` should support:

- `chat/completions` through OpenAI Chat Completions.
- `responses` through OpenAI Responses.
- Custom Base URL for OpenAI-compatible providers.
- Bearer API key configuration.

`AnthropicAdapter` should support:

- `messages` through Anthropic Messages.
- Native Anthropic API key configuration.
- Anthropic model listing if the SDK supports the required endpoint cleanly; otherwise model listing can be disabled with a friendly unsupported message.

Errors must not include API keys, request headers, or raw provider secrets.

## Prompt Templates

Prompt templates are fully owner-editable. The system should ship defaults, but it must not force hidden safety instructions after the owner customizes a template.

Supported operations:

- `improve`
- `shorten`
- `expand`
- `title_candidates`
- `summary_candidates`

Each operation should have an editable template. Templates can reference variables:

- `{{mode}}`
- `{{target}}`
- `{{text}}`
- `{{article_title}}`
- `{{article_summary}}`
- `{{article_excerpt}}`
- `{{locale}}`
- `{{max_suggestions}}`

The settings UI should allow:

- Editing each full template.
- Previewing the rendered prompt with sample data.
- Resetting one template to its built-in default.
- Resetting all templates to defaults.

Prompt templates should be stored server-side with AI config. A practical first implementation is adding a JSON column to `ai_provider_configs`, for example `prompt_templates jsonb NOT NULL DEFAULT '{}'`, then normalizing missing keys to built-in defaults at runtime.

## Output Parsing Contract

Built-in templates should ask for JSON:

```json
{
  "suggestions": [
    {
      "content": "候选内容",
      "reason": "可选原因"
    }
  ]
}
```

Parsing behavior:

- If provider output is valid JSON with non-empty `suggestions`, normalize and return up to `max_suggestions`.
- If provider output is valid JSON but has no usable suggestions, return the raw output as one candidate.
- If provider output is not JSON, return the trimmed raw output as one candidate.
- If provider output is empty, return a malformed response error.

This keeps full prompt freedom while avoiding the current failure mode where a non-JSON model response makes the feature unusable.

## Frontend UX

The editor gets a persistent AI candidate drawer:

- Desktop: right-side drawer inside the editor workbench, visually aligned with the article settings side area.
- Narrow screens: bottom drawer.
- The drawer shows the current operation, loading state, source snapshot, candidates, reasons, and actions.

Content selection flow:

1. Owner selects text in CKEditor.
2. Frontend captures selected plain text and a selection bookmark or equivalent range reference.
3. Owner clicks `润色`, `精简`, or `扩写`.
4. Drawer opens in loading state. Editor content stays unchanged.
5. Backend returns candidates.
6. Owner can copy, insert at current cursor, or replace the original selection.
7. Only `替换选区` restores the captured selection and writes content back to CKEditor.
8. Existing dirty-state and draft tracking mark the article as modified.

Title and summary flow:

1. Owner clicks `标题候选` or `摘要候选`.
2. Drawer opens with candidates.
3. Clicking a candidate can preview it, but not write to the field immediately.
4. Owner clicks `应用到标题` or `应用到摘要`.
5. Existing dirty-state and draft tracking mark the article as modified.

The UI should never save the article automatically after applying a candidate.

## Current Implementation Issues To Address

- The current `openAICompatiblePolisher` mixes multiple protocols in one implementation. Refactor to adapter classes.
- Prompt behavior is fixed in backend code. Move to editable templates with defaults.
- The frontend content application path inserts text at the current selection and may lose the original selected range after async calls. Capture a durable selection reference before calling AI.
- Candidate handling should support raw fallback output.
- The settings page should make provider/protocol/model/key and prompt templates feel like one coherent AI configuration surface.

## Error Handling

- Missing admin auth: existing admin auth behavior.
- Disabled or incomplete AI config: `FailedPrecondition`.
- Unsupported provider/protocol: `InvalidArgument` or `FailedPrecondition`, depending on whether it is a request or saved config issue.
- Provider timeout or upstream failure: `Unavailable`.
- Empty provider response: `Internal` mapped from malformed response.
- Non-JSON provider response: not an error; return raw fallback candidate.

Frontend failures show toast messages and keep content unchanged.

## Testing

Backend tests:

- Adapter selection by provider and protocol.
- OpenAI Chat Completions adapter request shape.
- OpenAI Responses adapter request shape.
- Anthropic Messages adapter request shape.
- Custom Base URL for OpenAI-compatible providers.
- Prompt template rendering and missing-variable behavior.
- Built-in template fallback for missing custom keys.
- JSON suggestions parsing.
- Raw fallback parsing.
- Secret redaction in errors.
- gRPC status mapping.
- AI config persistence for prompt templates.

Frontend tests:

- Request builder includes mode, target, selected text, and bounded context.
- Candidate drawer states: loading, success, empty, failure.
- Title and summary candidates require explicit apply.
- Content candidate replacement uses the captured original selection.
- Copy and insert actions do not overwrite the source selection.
- AI settings page renders prompt template editing and reset actions.

Verification should include backend unit tests, frontend unit/type tests, and a manual editor smoke test with a fake provider response.

## Implementation Defaults

- Anthropic model listing should show a friendly unsupported message in the first implementation unless the SDK exposes a clean model listing endpoint during implementation.
- Prompt template editing should live in an advanced section on the existing AI settings page, not in a separate route.
