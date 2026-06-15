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
	opts := []option.RequestOption{
		option.WithAPIKey(strings.TrimSpace(config.APIKey)),
		option.WithHTTPClient(newProviderHTTPClient(config)),
	}
	if baseURL := strings.TrimSpace(config.BaseURL); baseURL != "" {
		opts = append(opts, option.WithBaseURL(strings.TrimRight(baseURL, "/")))
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
