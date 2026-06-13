package gapi

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/internal/ai"
	"github.com/MonitorAllen/nostalgia/internal/secrets"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	aiPolishConfigPurpose = "ai_polish"
	aiPolishAPIKeyAAD     = "nostalgia:ai-polish-api-key"
	aiConfigSourceEnv     = "runtime_env"
	aiConfigSourceDB      = "database"
)

type resolvedAIConfig struct {
	Provider         string
	APIProtocol      string
	BaseURL          string
	Model            string
	APIKey           string
	Timeout          time.Duration
	MaxInputChars    int
	MaxContextChars  int
	MaxSuggestions   int
	EnabledRequested bool
	Source           string
}

func (server *Server) resolveAIPolishConfig(ctx context.Context) (resolvedAIConfig, error) {
	runtimeConfig := server.runtimeAIPolishConfig()
	if server.store == nil {
		return runtimeConfig, nil
	}

	row, err := server.store.GetAIProviderConfig(ctx, aiPolishConfigPurpose)
	if errors.Is(err, pgx.ErrNoRows) {
		return runtimeConfig, nil
	}
	if err != nil {
		return resolvedAIConfig{}, err
	}

	return server.aiConfigFromRow(row)
}

func (server *Server) runtimeAIPolishConfig() resolvedAIConfig {
	cfg := resolvedAIConfig{
		Provider:         strings.TrimSpace(server.config.AIPolishProvider),
		APIProtocol:      normalizeAIPolishProtocol(server.config.AIPolishAPIProtocol),
		BaseURL:          strings.TrimSpace(server.config.AIPolishBaseURL),
		Model:            strings.TrimSpace(server.config.AIPolishModel),
		APIKey:           strings.TrimSpace(server.config.AIPolishAPIKey),
		Timeout:          normalizedDuration(server.config.AIPolishTimeout, 30*time.Second),
		MaxInputChars:    normalizedPositive(server.config.AIPolishMaxInputChars, 6000),
		MaxContextChars:  normalizedNonNegative(server.config.AIPolishMaxContextChars, 4000),
		MaxSuggestions:   normalizedPositive(server.config.AIPolishMaxSuggestions, 3),
		EnabledRequested: true,
		Source:           aiConfigSourceEnv,
	}
	return cfg
}

func (server *Server) aiConfigFromRow(row db.AiProviderConfig) (resolvedAIConfig, error) {
	apiKey, err := secrets.DecryptString(row.ApiKeyCiphertext, server.config.TokenSymmetricKey, aiPolishAPIKeyAAD)
	if err != nil {
		apiKey = ""
	}

	return resolvedAIConfig{
		Provider:         strings.TrimSpace(row.Provider),
		APIProtocol:      normalizeAIPolishProtocol(row.ApiProtocol),
		BaseURL:          strings.TrimSpace(row.BaseUrl),
		Model:            strings.TrimSpace(row.Model),
		APIKey:           strings.TrimSpace(apiKey),
		Timeout:          normalizedDuration(time.Duration(row.TimeoutMs)*time.Millisecond, 30*time.Second),
		MaxInputChars:    normalizedPositive(int(row.MaxInputChars), 6000),
		MaxContextChars:  normalizedNonNegative(int(row.MaxContextChars), 4000),
		MaxSuggestions:   normalizedPositive(int(row.MaxSuggestions), 3),
		EnabledRequested: row.Enabled,
		Source:           aiConfigSourceDB,
	}, nil
}

func (cfg resolvedAIConfig) toResponse() *pb.GetAIConfigResponse {
	return &pb.GetAIConfigResponse{
		Provider:         cfg.Provider,
		ApiProtocol:      cfg.APIProtocol,
		BaseUrl:          cfg.BaseURL,
		Model:            cfg.Model,
		ApiKeyConfigured: cfg.apiKeyConfigured(),
		Enabled:          cfg.usable(),
		Timeout:          cfg.Timeout.String(),
		MaxInputChars:    int32(cfg.MaxInputChars),
		MaxContextChars:  int32(cfg.MaxContextChars),
		MaxSuggestions:   int32(cfg.MaxSuggestions),
		Source:           cfg.Source,
	}
}

func (cfg resolvedAIConfig) toRuntimeConfig(base util.Config) util.Config {
	base.AIPolishProvider = cfg.Provider
	base.AIPolishAPIProtocol = cfg.APIProtocol
	base.AIPolishBaseURL = cfg.BaseURL
	base.AIPolishAPIKey = cfg.APIKey
	base.AIPolishModel = cfg.Model
	base.AIPolishTimeout = cfg.Timeout
	base.AIPolishMaxInputChars = cfg.MaxInputChars
	base.AIPolishMaxContextChars = cfg.MaxContextChars
	base.AIPolishMaxSuggestions = cfg.MaxSuggestions
	return base
}

func (cfg resolvedAIConfig) apiKeyConfigured() bool {
	return strings.TrimSpace(cfg.APIKey) != ""
}

func (cfg resolvedAIConfig) usable() bool {
	return cfg.EnabledRequested &&
		strings.TrimSpace(cfg.Provider) != "" &&
		strings.TrimSpace(cfg.APIProtocol) != "" &&
		strings.TrimSpace(cfg.BaseURL) != "" &&
		strings.TrimSpace(cfg.Model) != "" &&
		cfg.apiKeyConfigured()
}

func (server *Server) buildUpdatedAIConfig(ctx context.Context, req *pb.UpdateAIConfigRequest) (resolvedAIConfig, error) {
	current, err := server.resolveAIPolishConfig(ctx)
	if err != nil {
		return resolvedAIConfig{}, err
	}

	provider := strings.TrimSpace(req.GetProvider())
	apiProtocol := current.APIProtocol
	if strings.TrimSpace(req.GetApiProtocol()) != "" {
		apiProtocol = normalizeAIPolishProtocol(req.GetApiProtocol())
	}
	baseURL := strings.TrimSpace(req.GetBaseUrl())
	model := strings.TrimSpace(req.GetModel())
	apiKey := current.APIKey
	if req.GetClearApiKey() {
		apiKey = ""
	} else if strings.TrimSpace(req.GetApiKey()) != "" {
		apiKey = strings.TrimSpace(req.GetApiKey())
	}

	timeout := current.Timeout
	if strings.TrimSpace(req.GetTimeout()) != "" {
		parsed, parseErr := time.ParseDuration(strings.TrimSpace(req.GetTimeout()))
		if parseErr != nil {
			return resolvedAIConfig{}, status.Error(codes.InvalidArgument, "invalid AI timeout")
		}
		timeout = parsed
	}

	cfg := resolvedAIConfig{
		Provider:         provider,
		APIProtocol:      apiProtocol,
		BaseURL:          baseURL,
		Model:            model,
		APIKey:           apiKey,
		Timeout:          timeout,
		MaxInputChars:    choosePositive(req.GetMaxInputChars(), current.MaxInputChars, 6000),
		MaxContextChars:  chooseNonNegative(req.GetMaxContextChars(), current.MaxContextChars, 4000),
		MaxSuggestions:   choosePositive(req.GetMaxSuggestions(), current.MaxSuggestions, 3),
		EnabledRequested: req.GetEnabled(),
		Source:           aiConfigSourceDB,
	}

	if err := validateUpdatedAIConfig(cfg); err != nil {
		return resolvedAIConfig{}, err
	}
	return cfg, nil
}

func validateUpdatedAIConfig(cfg resolvedAIConfig) error {
	if cfg.Provider == "" || len([]rune(cfg.Provider)) > 64 {
		return status.Error(codes.InvalidArgument, "AI provider name is required")
	}
	if !isSupportedAIPolishProtocol(cfg.APIProtocol) {
		return status.Error(codes.InvalidArgument, "unsupported AI API protocol")
	}
	parsedURL, err := url.Parse(cfg.BaseURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return status.Error(codes.InvalidArgument, "invalid AI base URL")
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return status.Error(codes.InvalidArgument, "invalid AI base URL")
	}
	if cfg.Model == "" {
		return status.Error(codes.InvalidArgument, "AI model is required")
	}
	if cfg.Timeout < time.Second || cfg.Timeout > 5*time.Minute {
		return status.Error(codes.InvalidArgument, "AI timeout must be between 1s and 5m")
	}
	if cfg.MaxInputChars < 1 || cfg.MaxInputChars > 20000 {
		return status.Error(codes.InvalidArgument, "AI max input chars must be between 1 and 20000")
	}
	if cfg.MaxContextChars < 0 || cfg.MaxContextChars > 20000 {
		return status.Error(codes.InvalidArgument, "AI max context chars must be between 0 and 20000")
	}
	if cfg.MaxSuggestions < 1 || cfg.MaxSuggestions > 5 {
		return status.Error(codes.InvalidArgument, "AI max suggestions must be between 1 and 5")
	}
	if cfg.EnabledRequested && !cfg.apiKeyConfigured() {
		return status.Error(codes.InvalidArgument, "AI API key is required when enabled")
	}
	return nil
}

func (server *Server) saveAIConfig(ctx context.Context, cfg resolvedAIConfig, updatedBy pgtype.UUID) (resolvedAIConfig, error) {
	ciphertext, err := secrets.EncryptString(cfg.APIKey, server.config.TokenSymmetricKey, aiPolishAPIKeyAAD)
	if err != nil {
		return resolvedAIConfig{}, err
	}

	row, err := server.store.UpsertAIProviderConfig(ctx, db.UpsertAIProviderConfigParams{
		Purpose:          aiPolishConfigPurpose,
		Provider:         cfg.Provider,
		ApiProtocol:      cfg.APIProtocol,
		BaseUrl:          cfg.BaseURL,
		Model:            cfg.Model,
		ApiKeyCiphertext: ciphertext,
		TimeoutMs:        int32(cfg.Timeout / time.Millisecond),
		MaxInputChars:    int32(cfg.MaxInputChars),
		MaxContextChars:  int32(cfg.MaxContextChars),
		MaxSuggestions:   int32(cfg.MaxSuggestions),
		Enabled:          cfg.EnabledRequested,
		UpdatedBy:        updatedBy,
	})
	if err != nil {
		return resolvedAIConfig{}, err
	}

	return server.aiConfigFromRow(row)
}

func normalizedDuration(value time.Duration, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return value
}

func normalizedPositive(value int, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func normalizedNonNegative(value int, fallback int) int {
	if value < 0 {
		return fallback
	}
	return value
}

func choosePositive(next int32, current int, fallback int) int {
	if next > 0 {
		return int(next)
	}
	return normalizedPositive(current, fallback)
}

func chooseNonNegative(next int32, current int, fallback int) int {
	if next >= 0 {
		return int(next)
	}
	return normalizedNonNegative(current, fallback)
}

func normalizeAIPolishProtocol(value string) string {
	switch strings.Trim(strings.TrimSpace(value), "/") {
	case ai.APIProtocolResponses:
		return ai.APIProtocolResponses
	case ai.APIProtocolMessages:
		return ai.APIProtocolMessages
	default:
		return ai.APIProtocolChatCompletions
	}
}

func isSupportedAIPolishProtocol(value string) bool {
	switch value {
	case ai.APIProtocolChatCompletions, ai.APIProtocolResponses, ai.APIProtocolMessages:
		return true
	default:
		return false
	}
}
