package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	ModeImprove           = "improve"
	ModeShorten           = "shorten"
	ModeExpand            = "expand"
	ModeTitleCandidates   = "title_candidates"
	ModeSummaryCandidates = "summary_candidates"

	TargetContentSelection = "content_selection"
	TargetTitle            = "title"
	TargetSummary          = "summary"

	APIProtocolChatCompletions = "chat/completions"
	APIProtocolResponses       = "responses"
	APIProtocolMessages        = "messages"

	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
)

var (
	ErrDisabled          = errors.New("ai polish is not configured")
	ErrInvalidInput      = errors.New("invalid ai polish input")
	ErrProviderFailure   = errors.New("ai provider failure")
	ErrMalformedResponse = errors.New("malformed ai provider response")
)

type TextPolisher interface {
	Polish(ctx context.Context, req PolishRequest) (PolishResponse, error)
}

type ModelLister interface {
	ListModels(ctx context.Context) ([]Model, error)
}

type Model struct {
	ID string
}

type ServiceConfig struct {
	Provider         string
	APIProtocol      string
	BaseURL          string
	APIKey           string
	Model            string
	Timeout          time.Duration
	MaxInputChars    int
	MaxContextChars  int
	MaxSuggestions   int
	PromptTemplates  map[string]string
	HTTPProxyAddress string
}

type GenerateRequest struct {
	Protocol string
	Model    string
	Prompt   string
}

type GenerateResponse struct {
	Content string
	Model   string
}

type ProviderAdapter interface {
	Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
	ListModels(ctx context.Context) ([]Model, error)
}

type ProviderFactory func(ServiceConfig) (ProviderAdapter, error)

type PolishRequest struct {
	Mode           string
	Target         string
	Text           string
	RichText       string
	InputFormat    string
	ArticleID      string
	ArticleTitle   string
	ArticleSummary string
	ArticleExcerpt string
	Locale         string
}

type Suggestion struct {
	Content string
	Reason  string
}

type PolishResponse struct {
	Suggestions []Suggestion
	Mode        string
	Target      string
	Model       string
}

func (req PolishRequest) normalized() PolishRequest {
	req.Mode = strings.TrimSpace(req.Mode)
	req.Target = strings.TrimSpace(req.Target)
	req.Text = strings.TrimSpace(req.Text)
	req.RichText = strings.TrimSpace(req.RichText)
	req.InputFormat = strings.TrimSpace(req.InputFormat)
	req.ArticleID = strings.TrimSpace(req.ArticleID)
	req.ArticleTitle = strings.TrimSpace(req.ArticleTitle)
	req.ArticleSummary = strings.TrimSpace(req.ArticleSummary)
	req.ArticleExcerpt = strings.TrimSpace(req.ArticleExcerpt)
	req.Locale = strings.TrimSpace(req.Locale)
	if req.Locale == "" {
		req.Locale = "zh-CN"
	}
	if req.InputFormat == "" {
		req.InputFormat = "plain_text"
	}
	return req
}

func validateRequest(req PolishRequest, maxInputChars int) error {
	req = req.normalized()
	if maxInputChars <= 0 {
		maxInputChars = 6000
	}
	if req.Mode == "" {
		return fmt.Errorf("%w: mode is required", ErrInvalidInput)
	}
	if req.Target == "" {
		return fmt.Errorf("%w: target is required", ErrInvalidInput)
	}
	switch req.Mode {
	case ModeImprove, ModeShorten, ModeExpand:
		if req.Target != TargetContentSelection {
			return fmt.Errorf("%w: mode %s requires content selection target", ErrInvalidInput, req.Mode)
		}
		if req.Text == "" {
			return fmt.Errorf("%w: text is required", ErrInvalidInput)
		}
	case ModeTitleCandidates:
		if req.Target != TargetTitle {
			return fmt.Errorf("%w: title candidates require title target", ErrInvalidInput)
		}
	case ModeSummaryCandidates:
		if req.Target != TargetSummary {
			return fmt.Errorf("%w: summary candidates require summary target", ErrInvalidInput)
		}
	default:
		return fmt.Errorf("%w: unsupported mode", ErrInvalidInput)
	}
	if len([]rune(req.Text)) > maxInputChars {
		return fmt.Errorf("%w: text exceeds maximum length", ErrInvalidInput)
	}
	return nil
}

func ValidateRequest(req PolishRequest, maxInputChars int) error {
	return validateRequest(req, maxInputChars)
}

func limitRunes(value string, max int) string {
	if max <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}
