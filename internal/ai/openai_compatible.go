package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	apiProtocol     string
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

type responsesRequest struct {
	Model       string        `json:"model"`
	Input       []ChatMessage `json:"input"`
	Temperature float64       `json:"temperature"`
}

type responsesResponse struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

type messagesRequest struct {
	Model     string        `json:"model"`
	System    string        `json:"system,omitempty"`
	Messages  []ChatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens"`
}

type messagesResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

type modelListResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

type suggestionEnvelope struct {
	Suggestions []Suggestion `json:"suggestions"`
}

func NewOpenAICompatiblePolisher(config util.Config) TextPolisher {
	return &openAICompatiblePolisher{
		baseURL:         strings.TrimSpace(config.AIPolishBaseURL),
		apiKey:          strings.TrimSpace(config.AIPolishAPIKey),
		model:           strings.TrimSpace(config.AIPolishModel),
		apiProtocol:     normalizeAPIProtocol(config.AIPolishAPIProtocol),
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

	body, err := json.Marshal(polisher.requestPayload(BuildMessages(req)))
	if err != nil {
		return PolishResponse{}, fmt.Errorf("%w: failed to encode request", ErrProviderFailure)
	}

	httpReq, err := http.NewRequestWithContext(callCtx, http.MethodPost, polisher.endpoint(), bytes.NewReader(body))
	if err != nil {
		return PolishResponse{}, fmt.Errorf("%w: failed to build request", ErrProviderFailure)
	}
	polisher.applyHeaders(httpReq)

	resp, err := polisher.httpClient().Do(httpReq)
	if err != nil {
		return PolishResponse{}, fmt.Errorf("%w: request failed", ErrProviderFailure)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return PolishResponse{}, fmt.Errorf("%w: upstream status %d", ErrProviderFailure, resp.StatusCode)
	}

	content, err := polisher.decodeContent(resp.Body)
	if err != nil {
		return PolishResponse{}, err
	}

	var envelope suggestionEnvelope
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &envelope); err != nil {
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
	return strings.TrimRight(polisher.baseURL, "/") + "/" + polisher.apiProtocol
}

func (polisher *openAICompatiblePolisher) modelsEndpoint() string {
	return strings.TrimRight(polisher.baseURL, "/") + "/models"
}

func (polisher *openAICompatiblePolisher) requestPayload(messages []ChatMessage) any {
	switch polisher.apiProtocol {
	case APIProtocolResponses:
		return responsesRequest{
			Model:       polisher.model,
			Input:       messages,
			Temperature: 0.2,
		}
	case APIProtocolMessages:
		userMessages := messages
		system := ""
		if len(messages) > 0 && messages[0].Role == "system" {
			system = messages[0].Content
			userMessages = messages[1:]
		}
		return messagesRequest{
			Model:     polisher.model,
			System:    system,
			Messages:  userMessages,
			MaxTokens: 2048,
		}
	default:
		return chatCompletionRequest{
			Model:       polisher.model,
			Messages:    messages,
			Temperature: 0.2,
			ResponseFormat: map[string]string{
				"type": "json_object",
			},
		}
	}
}

func (polisher *openAICompatiblePolisher) decodeContent(body io.Reader) (string, error) {
	switch polisher.apiProtocol {
	case APIProtocolResponses:
		var providerResp responsesResponse
		if err := json.NewDecoder(body).Decode(&providerResp); err != nil {
			return "", fmt.Errorf("%w: invalid provider json", ErrMalformedResponse)
		}
		content := strings.TrimSpace(providerResp.OutputText)
		if content == "" {
			var parts []string
			for _, output := range providerResp.Output {
				for _, item := range output.Content {
					if strings.TrimSpace(item.Text) != "" {
						parts = append(parts, item.Text)
					}
				}
			}
			content = strings.TrimSpace(strings.Join(parts, "\n"))
		}
		if content == "" {
			return "", fmt.Errorf("%w: missing provider content", ErrMalformedResponse)
		}
		return content, nil
	case APIProtocolMessages:
		var providerResp messagesResponse
		if err := json.NewDecoder(body).Decode(&providerResp); err != nil {
			return "", fmt.Errorf("%w: invalid provider json", ErrMalformedResponse)
		}
		var parts []string
		for _, item := range providerResp.Content {
			if strings.TrimSpace(item.Text) != "" {
				parts = append(parts, item.Text)
			}
		}
		content := strings.TrimSpace(strings.Join(parts, "\n"))
		if content == "" {
			return "", fmt.Errorf("%w: missing provider content", ErrMalformedResponse)
		}
		return content, nil
	default:
		var providerResp chatCompletionResponse
		if err := json.NewDecoder(body).Decode(&providerResp); err != nil {
			return "", fmt.Errorf("%w: invalid provider json", ErrMalformedResponse)
		}
		if len(providerResp.Choices) == 0 || strings.TrimSpace(providerResp.Choices[0].Message.Content) == "" {
			return "", fmt.Errorf("%w: missing provider content", ErrMalformedResponse)
		}
		return strings.TrimSpace(providerResp.Choices[0].Message.Content), nil
	}
}

func (polisher *openAICompatiblePolisher) applyHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if polisher.apiProtocol == APIProtocolMessages {
		req.Header.Set("x-api-key", polisher.apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
		return
	}
	req.Header.Set("Authorization", "Bearer "+polisher.apiKey)
}

func (polisher *openAICompatiblePolisher) ListModels(ctx context.Context) ([]Model, error) {
	if polisher.baseURL == "" || polisher.apiKey == "" {
		return nil, ErrDisabled
	}

	timeout := polisher.normalizedTimeout()
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(callCtx, http.MethodGet, polisher.modelsEndpoint(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to build models request", ErrProviderFailure)
	}
	polisher.applyHeaders(httpReq)

	resp, err := polisher.httpClient().Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: models request failed", ErrProviderFailure)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%w: upstream models status %d", ErrProviderFailure, resp.StatusCode)
	}

	var providerResp modelListResponse
	if err := json.NewDecoder(resp.Body).Decode(&providerResp); err != nil {
		return nil, fmt.Errorf("%w: invalid models json", ErrMalformedResponse)
	}

	models := make([]Model, 0, len(providerResp.Data))
	for _, item := range providerResp.Data {
		if id := strings.TrimSpace(item.ID); id != "" {
			models = append(models, Model{ID: id})
		}
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("%w: empty models list", ErrMalformedResponse)
	}
	return models, nil
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
		Timeout:   polisher.normalizedTimeout(),
	}
}

func (polisher *openAICompatiblePolisher) normalizedTimeout() time.Duration {
	if polisher.timeout <= 0 {
		return 30 * time.Second
	}
	return polisher.timeout
}

func normalizeAPIProtocol(value string) string {
	switch strings.Trim(strings.TrimSpace(value), "/") {
	case APIProtocolResponses:
		return APIProtocolResponses
	case APIProtocolMessages:
		return APIProtocolMessages
	default:
		return APIProtocolChatCompletions
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
