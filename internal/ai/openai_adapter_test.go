package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAIAdapterUsesChatCompletions(t *testing.T) {
	var gotPath string
	var gotPrompt string
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		require.Equal(t, "Bearer secret-key", r.Header.Get("Authorization"))
		var payload struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "writer-model", payload.Model)
		require.Len(t, payload.Messages, 1)
		gotPrompt = payload.Messages[0].Content

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl_1","object":"chat.completion","created":0,"model":"writer-model","choices":[{"index":0,"message":{"role":"assistant","content":"raw candidate"},"finish_reason":"stop"}]}`))
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
	require.Equal(t, "hello", gotPrompt)
	require.Equal(t, "raw candidate", resp.Content)
}

func TestOpenAIAdapterUsesResponses(t *testing.T) {
	var gotPath string
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		require.Equal(t, "Bearer secret-key", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"resp_1","object":"response","created_at":0,"status":"completed","model":"writer-model","output":[{"type":"message","id":"msg_1","status":"completed","role":"assistant","content":[{"type":"output_text","text":"responses candidate","annotations":[]}]}]}`))
	}))
	defer provider.Close()

	adapter := NewOpenAIAdapter(ServiceConfig{
		BaseURL:     provider.URL,
		APIKey:      "secret-key",
		APIProtocol: APIProtocolResponses,
		Model:       "writer-model",
	})

	resp, err := adapter.Generate(context.Background(), GenerateRequest{
		Protocol: APIProtocolResponses,
		Model:    "writer-model",
		Prompt:   "hello",
	})

	require.NoError(t, err)
	require.Equal(t, "/responses", gotPath)
	require.Equal(t, "responses candidate", resp.Content)
}

func TestOpenAIAdapterListsModels(t *testing.T) {
	var gotPath string
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		require.Equal(t, "Bearer secret-key", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"writer-model","object":"model","created":0,"owned_by":"nostalgia"},{"id":"fast-model","object":"model","created":0,"owned_by":"nostalgia"}]}`))
	}))
	defer provider.Close()

	adapter := NewOpenAIAdapter(ServiceConfig{
		BaseURL: provider.URL,
		APIKey:  "secret-key",
		Model:   "writer-model",
	})

	models, err := adapter.ListModels(context.Background())

	require.NoError(t, err)
	require.Equal(t, "/models", gotPath)
	require.Equal(t, []Model{{ID: "writer-model"}, {ID: "fast-model"}}, models)
}
