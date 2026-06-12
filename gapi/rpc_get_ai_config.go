package gapi

import (
	"context"
	"strings"

	"github.com/MonitorAllen/nostalgia/pb"
)

func (server *Server) GetAIConfig(ctx context.Context, _ *pb.GetAIConfigRequest) (*pb.GetAIConfigResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	provider := strings.TrimSpace(server.config.AIPolishProvider)
	baseURL := strings.TrimSpace(server.config.AIPolishBaseURL)
	model := strings.TrimSpace(server.config.AIPolishModel)
	apiKeyConfigured := strings.TrimSpace(server.config.AIPolishAPIKey) != ""

	return &pb.GetAIConfigResponse{
		Provider:         provider,
		BaseUrl:          baseURL,
		Model:            model,
		ApiKeyConfigured: apiKeyConfigured,
		Enabled:          provider != "" && baseURL != "" && model != "" && apiKeyConfigured,
		Timeout:          server.config.AIPolishTimeout.String(),
		MaxInputChars:    int32(server.config.AIPolishMaxInputChars),
		MaxContextChars:  int32(server.config.AIPolishMaxContextChars),
		MaxSuggestions:   int32(server.config.AIPolishMaxSuggestions),
		Source:           "runtime_env",
	}, nil
}
