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
		BaseURL:         "https://ai.example.com/v1",
		APIKey:          "secret-key",
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

func TestPolishServiceParsesMarkdownFencedJSONForShorten(t *testing.T) {
	adapter := &fakeProviderAdapter{output: "```json\n" +
		`{"suggestions":[{"content":"更短","reason":"去掉冗余"}]}` +
		"\n```"}
	service := NewPolishService(ServiceConfig{
		Provider:        "openai",
		APIProtocol:     APIProtocolChatCompletions,
		BaseURL:         "https://ai.example.com/v1",
		APIKey:          "secret-key",
		Model:           "writer-model",
		MaxInputChars:   6000,
		MaxContextChars: 4000,
		MaxSuggestions:  3,
		PromptTemplates: map[string]string{ModeShorten: "shorten {{text}}"},
	}, func(ServiceConfig) (ProviderAdapter, error) {
		return adapter, nil
	})

	resp, err := service.Polish(context.Background(), PolishRequest{
		Mode:   ModeShorten,
		Target: TargetContentSelection,
		Text:   "需要精简的原文",
	})

	require.NoError(t, err)
	require.Equal(t, []Suggestion{{Content: "更短", Reason: "去掉冗余"}}, resp.Suggestions)
}

func TestPolishServiceRendersRichTextSelectionContext(t *testing.T) {
	adapter := &fakeProviderAdapter{output: `{"suggestions":[{"content":"<ul><li><strong>更清晰</strong></li></ul>"}]}`}
	service := NewPolishService(ServiceConfig{
		Provider:       "openai",
		APIProtocol:    APIProtocolChatCompletions,
		BaseURL:        "https://ai.example.com/v1",
		APIKey:         "secret-key",
		Model:          "writer-model",
		MaxInputChars:  6000,
		MaxSuggestions: 3,
		PromptTemplates: map[string]string{
			ModeImprove: "format={{input_format}}\nplain={{text}}\nrich={{rich_text}}",
		},
	}, func(ServiceConfig) (ProviderAdapter, error) {
		return adapter, nil
	})

	resp, err := service.Polish(context.Background(), PolishRequest{
		Mode:        ModeImprove,
		Target:      TargetContentSelection,
		Text:        "推荐的配置分层",
		RichText:    "<h2>推荐的配置分层</h2><ul><li>本地开发</li></ul>",
		InputFormat: "html",
	})

	require.NoError(t, err)
	require.Contains(t, adapter.request.Prompt, "format=html")
	require.Contains(t, adapter.request.Prompt, "plain=推荐的配置分层")
	require.Contains(t, adapter.request.Prompt, "rich=<h2>推荐的配置分层</h2><ul><li>本地开发</li></ul>")
	require.Equal(t, "<ul><li><strong>更清晰</strong></li></ul>", resp.Suggestions[0].Content)
}

func TestPolishServiceFallsBackToRawCandidate(t *testing.T) {
	adapter := &fakeProviderAdapter{output: "直接返回文本"}
	service := NewPolishService(ServiceConfig{
		Provider:       "openai",
		APIProtocol:    APIProtocolResponses,
		BaseURL:        "https://ai.example.com/v1",
		APIKey:         "secret-key",
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

func TestPolishServiceDisabledWhenConfigIncomplete(t *testing.T) {
	service := NewPolishService(ServiceConfig{
		Provider: "openai",
		BaseURL:  "https://ai.example.com/v1",
		Model:    "writer-model",
	}, nil)

	_, err := service.Polish(context.Background(), PolishRequest{
		Mode:   ModeImprove,
		Target: TargetContentSelection,
		Text:   "hello",
	})

	require.ErrorIs(t, err, ErrDisabled)
}

func TestPolishServiceValidatesInput(t *testing.T) {
	adapter := &fakeProviderAdapter{output: `{"suggestions":[{"content":"ok"}]}`}
	service := NewPolishService(ServiceConfig{
		Provider:       "openai",
		APIProtocol:    APIProtocolChatCompletions,
		BaseURL:        "https://ai.example.com/v1",
		APIKey:         "secret-key",
		Model:          "writer-model",
		MaxInputChars:  4,
		MaxSuggestions: 3,
	}, func(ServiceConfig) (ProviderAdapter, error) {
		return adapter, nil
	})

	_, err := service.Polish(context.Background(), PolishRequest{
		Mode:   ModeImprove,
		Target: TargetContentSelection,
		Text:   "hello",
	})

	require.ErrorIs(t, err, ErrInvalidInput)
}
