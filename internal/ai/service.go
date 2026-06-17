package ai

import (
	"context"
	"fmt"
	"strings"
)

type PolishService struct {
	config  ServiceConfig
	factory ProviderFactory
}

func NewPolishService(config ServiceConfig, factory ProviderFactory) TextPolisher {
	if factory == nil {
		factory = NewProviderAdapter
	}
	config.Provider = normalizeProvider(config.Provider)
	config.APIProtocol = normalizeProviderAPIProtocol(config.APIProtocol)
	config.BaseURL = strings.TrimSpace(config.BaseURL)
	config.APIKey = strings.TrimSpace(config.APIKey)
	config.Model = strings.TrimSpace(config.Model)
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
	req.RichText = limitRunes(req.RichText, service.config.MaxInputChars)

	adapter, err := service.factory(service.config)
	if err != nil {
		return PolishResponse{}, err
	}

	prompt := RenderPromptTemplate(service.config.PromptTemplates[req.Mode], PromptRenderData{
		Mode:           req.Mode,
		Target:         req.Target,
		Text:           req.Text,
		RichText:       req.RichText,
		InputFormat:    req.InputFormat,
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

func normalizeProvider(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "anthropic", "claude":
		return ProviderAnthropic
	default:
		return ProviderOpenAI
	}
}

func IsAnthropicProvider(value string) bool {
	return normalizeProvider(value) == ProviderAnthropic
}

func normalizeProviderAPIProtocol(value string) string {
	switch strings.Trim(strings.TrimSpace(value), "/") {
	case APIProtocolResponses:
		return APIProtocolResponses
	case APIProtocolMessages:
		return APIProtocolMessages
	default:
		return APIProtocolChatCompletions
	}
}
