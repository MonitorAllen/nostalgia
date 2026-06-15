package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnthropicAdapterUsesMessages(t *testing.T) {
	var gotPath string
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		require.Equal(t, "secret-key", r.Header.Get("x-api-key"))
		w.Header().Set("Content-Type", "application/json")
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
