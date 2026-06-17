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
	opts := []option.RequestOption{
		option.WithAPIKey(strings.TrimSpace(config.APIKey)),
		option.WithHTTPClient(newProviderHTTPClient(config)),
	}
	if baseURL := strings.TrimSpace(config.BaseURL); baseURL != "" {
		opts = append(opts, option.WithBaseURL(strings.TrimRight(baseURL, "/")))
	}
	return &OpenAIAdapter{config: config, client: openai.NewClient(opts...)}
}

func (adapter *OpenAIAdapter) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	protocol := normalizeProviderAPIProtocol(req.Protocol)
	switch protocol {
	case APIProtocolResponses:
		resp, err := adapter.client.Responses.New(ctx, responses.ResponseNewParams{
			Model: openai.ResponsesModel(req.Model),
			Input: responses.ResponseNewParamsInputUnion{
				OfString: openai.String(req.Prompt),
			},
		})
		if err != nil {
			return GenerateResponse{}, fmt.Errorf("%w: openai responses request failed", ErrProviderFailure)
		}
		content := strings.TrimSpace(resp.OutputText())
		if content == "" {
			return GenerateResponse{}, fmt.Errorf("%w: missing openai response content", ErrMalformedResponse)
		}
		return GenerateResponse{Content: content, Model: req.Model}, nil
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
		content := strings.TrimSpace(resp.Choices[0].Message.Content)
		if content == "" {
			return GenerateResponse{}, fmt.Errorf("%w: missing openai message content", ErrMalformedResponse)
		}
		return GenerateResponse{Content: content, Model: req.Model}, nil
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
