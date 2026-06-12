package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MonitorAllen/nostalgia/util"
)

type openAICompatiblePolisher struct {
	baseURL         string
	apiKey          string
	model           string
	timeout         time.Duration
	maxInputChars   int
	maxContextChars int
	maxSuggestions  int
	httpProxyAddr   string
}

type chatCompletionRequest struct {
	Model          string            `json:"model"`
	Messages       []ChatMessage     `json:"messages"`
	Temperature    float64           `json:"temperature"`
	ResponseFormat map[string]string `json:"response_format,omitempty"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type suggestionEnvelope struct {
	Suggestions []Suggestion `json:"suggestions"`
}

func NewOpenAICompatiblePolisher(config util.Config) TextPolisher {
	return &openAICompatiblePolisher{
		baseURL:         strings.TrimSpace(config.AIPolishBaseURL),
		apiKey:          strings.TrimSpace(config.AIPolishAPIKey),
		model:           strings.TrimSpace(config.AIPolishModel),
		timeout:         config.AIPolishTimeout,
		maxInputChars:   config.AIPolishMaxInputChars,
		maxContextChars: config.AIPolishMaxContextChars,
		maxSuggestions:  config.AIPolishMaxSuggestions,
		httpProxyAddr:   strings.TrimSpace(config.HTTPProxyAddr),
	}
}

func (polisher *openAICompatiblePolisher) Polish(ctx context.Context, req PolishRequest) (PolishResponse, error) {
	req = req.normalized()
	if polisher.disabled() {
		return PolishResponse{}, ErrDisabled
	}
	if err := validateRequest(req, polisher.maxInputChars); err != nil {
		return PolishResponse{}, err
	}

	req.ArticleTitle = limitRunes(req.ArticleTitle, polisher.maxContextChars)
	req.ArticleSummary = limitRunes(req.ArticleSummary, polisher.maxContextChars)
	req.ArticleExcerpt = limitRunes(req.ArticleExcerpt, polisher.maxContextChars)

	timeout := polisher.timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	payload := chatCompletionRequest{
		Model:       polisher.model,
		Messages:    BuildMessages(req),
		Temperature: 0.2,
		ResponseFormat: map[string]string{
			"type": "json_object",
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return PolishResponse{}, fmt.Errorf("%w: failed to encode request", ErrProviderFailure)
	}

	httpReq, err := http.NewRequestWithContext(callCtx, http.MethodPost, polisher.endpoint(), bytes.NewReader(body))
	if err != nil {
		return PolishResponse{}, fmt.Errorf("%w: failed to build request", ErrProviderFailure)
	}
	httpReq.Header.Set("Authorization", "Bearer "+polisher.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := polisher.httpClient().Do(httpReq)
	if err != nil {
		return PolishResponse{}, fmt.Errorf("%w: request failed", ErrProviderFailure)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return PolishResponse{}, fmt.Errorf("%w: upstream status %d", ErrProviderFailure, resp.StatusCode)
	}

	var providerResp chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&providerResp); err != nil {
		return PolishResponse{}, fmt.Errorf("%w: invalid provider json", ErrMalformedResponse)
	}
	if len(providerResp.Choices) == 0 || strings.TrimSpace(providerResp.Choices[0].Message.Content) == "" {
		return PolishResponse{}, fmt.Errorf("%w: missing provider content", ErrMalformedResponse)
	}

	var envelope suggestionEnvelope
	if err := json.Unmarshal([]byte(strings.TrimSpace(providerResp.Choices[0].Message.Content)), &envelope); err != nil {
		return PolishResponse{}, fmt.Errorf("%w: invalid suggestion json", ErrMalformedResponse)
	}

	suggestions := normalizeSuggestions(envelope.Suggestions, polisher.maxSuggestions)
	if len(suggestions) == 0 {
		return PolishResponse{}, fmt.Errorf("%w: no suggestions", ErrMalformedResponse)
	}

	return PolishResponse{
		Suggestions: suggestions,
		Mode:        req.Mode,
		Target:      req.Target,
		Model:       polisher.model,
	}, nil
}

func (polisher *openAICompatiblePolisher) disabled() bool {
	return polisher.baseURL == "" || polisher.apiKey == "" || polisher.model == ""
}

func (polisher *openAICompatiblePolisher) endpoint() string {
	return strings.TrimRight(polisher.baseURL, "/") + "/chat/completions"
}

func (polisher *openAICompatiblePolisher) httpClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if polisher.httpProxyAddr != "" {
		if proxyURL, err := url.Parse(polisher.httpProxyAddr); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}
	return &http.Client{
		Transport: transport,
		Timeout:   polisher.timeout,
	}
}

func normalizeSuggestions(values []Suggestion, max int) []Suggestion {
	if max <= 0 {
		max = 3
	}
	suggestions := make([]Suggestion, 0, min(len(values), max))
	for _, suggestion := range values {
		content := strings.TrimSpace(suggestion.Content)
		if content == "" {
			continue
		}
		suggestions = append(suggestions, Suggestion{
			Content: content,
			Reason:  strings.TrimSpace(suggestion.Reason),
		})
		if len(suggestions) == max {
			break
		}
	}
	return suggestions
}
