package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MonitorAllen/nostalgia/util"
	"github.com/stretchr/testify/require"
)

func TestOpenAICompatiblePolisherParsesSuggestions(t *testing.T) {
	var gotPath string
	var gotAuth string
	var gotModel string
	var gotMessages []map[string]string

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")

		var payload struct {
			Model    string              `json:"model"`
			Messages []map[string]string `json:"messages"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		gotModel = payload.Model
		gotMessages = payload.Messages

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [
				{
					"message": {
						"content": "{\"suggestions\":[{\"content\":\"更自然的表达\",\"reason\":\"语气更顺\"}]}"
					}
				}
			]
		}`))
	}))
	defer provider.Close()

	polisher := NewOpenAICompatiblePolisher(util.Config{
		AIPolishBaseURL:         provider.URL + "/v1/",
		AIPolishAPIKey:          "secret-key",
		AIPolishModel:           "writer-model",
		AIPolishTimeout:         time.Second,
		AIPolishMaxInputChars:   6000,
		AIPolishMaxContextChars: 4000,
		AIPolishMaxSuggestions:  3,
	})

	result, err := polisher.Polish(context.Background(), PolishRequest{
		Mode:           ModeImprove,
		Target:         TargetContentSelection,
		Text:           "原始表达",
		ArticleTitle:   "缓存策略",
		ArticleSummary: "介绍缓存失效",
		Locale:         "zh-CN",
	})
	require.NoError(t, err)
	require.Equal(t, "/v1/chat/completions", gotPath)
	require.Equal(t, "Bearer secret-key", gotAuth)
	require.Equal(t, "writer-model", gotModel)
	require.NotEmpty(t, gotMessages)
	require.Len(t, result.Suggestions, 1)
	require.Equal(t, "更自然的表达", result.Suggestions[0].Content)
	require.Equal(t, "语气更顺", result.Suggestions[0].Reason)
	require.Equal(t, "writer-model", result.Model)
}

func TestOpenAICompatiblePolisherUsesResponsesProtocol(t *testing.T) {
	var gotPath string
	var gotModel string
	var gotInput []map[string]string

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path

		var payload struct {
			Model string              `json:"model"`
			Input []map[string]string `json:"input"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		gotModel = payload.Model
		gotInput = payload.Input

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"output_text": "{\"suggestions\":[{\"content\":\"Responses 表达\",\"reason\":\"协议正确\"}]}"
		}`))
	}))
	defer provider.Close()

	polisher := NewOpenAICompatiblePolisher(util.Config{
		AIPolishBaseURL:         provider.URL + "/v1/",
		AIPolishAPIKey:          "secret-key",
		AIPolishModel:           "writer-model",
		AIPolishAPIProtocol:     APIProtocolResponses,
		AIPolishTimeout:         time.Second,
		AIPolishMaxInputChars:   6000,
		AIPolishMaxContextChars: 4000,
		AIPolishMaxSuggestions:  3,
	})

	result, err := polisher.Polish(context.Background(), PolishRequest{
		Mode:   ModeImprove,
		Target: TargetContentSelection,
		Text:   "原始表达",
	})

	require.NoError(t, err)
	require.Equal(t, "/v1/responses", gotPath)
	require.Equal(t, "writer-model", gotModel)
	require.Len(t, gotInput, 2)
	require.Equal(t, "Responses 表达", result.Suggestions[0].Content)
}

func TestOpenAICompatiblePolisherUsesMessagesProtocol(t *testing.T) {
	var gotPath string
	var gotAPIKey string
	var gotVersion string
	var gotModel string
	var gotMessages []map[string]string

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAPIKey = r.Header.Get("x-api-key")
		gotVersion = r.Header.Get("anthropic-version")

		var payload struct {
			Model    string              `json:"model"`
			Messages []map[string]string `json:"messages"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		gotModel = payload.Model
		gotMessages = payload.Messages

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"content": [
				{"type": "text", "text": "{\"suggestions\":[{\"content\":\"Messages 表达\",\"reason\":\"协议正确\"}]}"}
			]
		}`))
	}))
	defer provider.Close()

	polisher := NewOpenAICompatiblePolisher(util.Config{
		AIPolishBaseURL:         provider.URL + "/v1/",
		AIPolishAPIKey:          "secret-key",
		AIPolishModel:           "writer-model",
		AIPolishAPIProtocol:     APIProtocolMessages,
		AIPolishTimeout:         time.Second,
		AIPolishMaxInputChars:   6000,
		AIPolishMaxContextChars: 4000,
		AIPolishMaxSuggestions:  3,
	})

	result, err := polisher.Polish(context.Background(), PolishRequest{
		Mode:   ModeImprove,
		Target: TargetContentSelection,
		Text:   "原始表达",
	})

	require.NoError(t, err)
	require.Equal(t, "/v1/messages", gotPath)
	require.Equal(t, "secret-key", gotAPIKey)
	require.NotEmpty(t, gotVersion)
	require.Equal(t, "writer-model", gotModel)
	require.Len(t, gotMessages, 1)
	require.Equal(t, "Messages 表达", result.Suggestions[0].Content)
}

func TestOpenAICompatiblePolisherListsModelsFromStandardEndpoint(t *testing.T) {
	var gotPath string
	var gotAuth string

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": [
				{"id": "writer-model"},
				{"id": "fast-model"}
			]
		}`))
	}))
	defer provider.Close()

	polisher := NewOpenAICompatiblePolisher(util.Config{
		AIPolishBaseURL:     provider.URL + "/v1/",
		AIPolishAPIKey:      "secret-key",
		AIPolishModel:       "writer-model",
		AIPolishAPIProtocol: APIProtocolChatCompletions,
		AIPolishTimeout:     time.Second,
	})
	lister, ok := polisher.(ModelLister)
	require.True(t, ok)

	models, err := lister.ListModels(context.Background())

	require.NoError(t, err)
	require.Equal(t, "/v1/models", gotPath)
	require.Equal(t, "Bearer secret-key", gotAuth)
	require.Equal(t, []Model{{ID: "writer-model"}, {ID: "fast-model"}}, models)
}

func TestOpenAICompatiblePolisherDisabledWhenConfigIncomplete(t *testing.T) {
	polisher := NewOpenAICompatiblePolisher(util.Config{
		AIPolishBaseURL: "https://ai.example.com/v1",
		AIPolishModel:   "writer-model",
	})

	_, err := polisher.Polish(context.Background(), PolishRequest{
		Mode:   ModeImprove,
		Target: TargetContentSelection,
		Text:   "hello",
	})
	require.ErrorIs(t, err, ErrDisabled)
}

func TestOpenAICompatiblePolisherValidatesInput(t *testing.T) {
	polisher := NewOpenAICompatiblePolisher(util.Config{
		AIPolishBaseURL:        "https://ai.example.com/v1",
		AIPolishAPIKey:         "secret-key",
		AIPolishModel:          "writer-model",
		AIPolishMaxInputChars:  4,
		AIPolishMaxSuggestions: 3,
	})

	_, err := polisher.Polish(context.Background(), PolishRequest{
		Mode:   ModeImprove,
		Target: TargetContentSelection,
		Text:   "hello",
	})
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestOpenAICompatiblePolisherMapsProviderFailureWithoutLeakingSecret(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream includes secret-key", http.StatusBadGateway)
	}))
	defer provider.Close()

	polisher := NewOpenAICompatiblePolisher(util.Config{
		AIPolishBaseURL:        provider.URL,
		AIPolishAPIKey:         "secret-key",
		AIPolishModel:          "writer-model",
		AIPolishTimeout:        time.Second,
		AIPolishMaxInputChars:  6000,
		AIPolishMaxSuggestions: 3,
	})

	_, err := polisher.Polish(context.Background(), PolishRequest{
		Mode:   ModeImprove,
		Target: TargetContentSelection,
		Text:   "hello",
	})
	require.ErrorIs(t, err, ErrProviderFailure)
	require.NotContains(t, err.Error(), "secret-key")
}

func TestOpenAICompatiblePolisherMapsMalformedResponse(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"not-json"}}]}`))
	}))
	defer provider.Close()

	polisher := NewOpenAICompatiblePolisher(util.Config{
		AIPolishBaseURL:        provider.URL,
		AIPolishAPIKey:         "secret-key",
		AIPolishModel:          "writer-model",
		AIPolishTimeout:        time.Second,
		AIPolishMaxInputChars:  6000,
		AIPolishMaxSuggestions: 3,
	})

	_, err := polisher.Polish(context.Background(), PolishRequest{
		Mode:   ModeImprove,
		Target: TargetContentSelection,
		Text:   "hello",
	})
	require.ErrorIs(t, err, ErrMalformedResponse)
}

func TestBuildPromptKeepsModeSpecificInstructions(t *testing.T) {
	messages := BuildMessages(PolishRequest{
		Mode:           ModeShorten,
		Target:         TargetContentSelection,
		Text:           "请帮我把这段话写得更短。",
		ArticleTitle:   "文章标题",
		ArticleSummary: "文章摘要",
		ArticleExcerpt: strings.Repeat("正文", 10),
		Locale:         "zh-CN",
	})

	require.Len(t, messages, 2)
	require.Equal(t, "system", messages[0].Role)
	require.Contains(t, messages[0].Content, "strict JSON")
	require.Equal(t, "user", messages[1].Role)
	require.Contains(t, messages[1].Content, "shorten")
	require.Contains(t, messages[1].Content, "文章标题")
}
