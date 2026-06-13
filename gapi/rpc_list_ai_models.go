package gapi

import (
	"context"
	"errors"
	"strings"

	"github.com/MonitorAllen/nostalgia/internal/ai"
	"github.com/MonitorAllen/nostalgia/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) ListAIModels(ctx context.Context, req *pb.ListAIModelsRequest) (*pb.ListAIModelsResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	cfg, err := server.resolveAIPolishConfig(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to load AI config")
	}

	if strings.TrimSpace(req.GetProvider()) != "" {
		cfg.Provider = strings.TrimSpace(req.GetProvider())
	}
	if strings.TrimSpace(req.GetApiProtocol()) != "" {
		cfg.APIProtocol = normalizeAIPolishProtocol(req.GetApiProtocol())
	}
	if strings.TrimSpace(req.GetBaseUrl()) != "" {
		cfg.BaseURL = strings.TrimSpace(req.GetBaseUrl())
	}
	if strings.TrimSpace(req.GetApiKey()) != "" {
		cfg.APIKey = strings.TrimSpace(req.GetApiKey())
	}

	if cfg.BaseURL == "" || cfg.APIKey == "" {
		return nil, status.Error(codes.FailedPrecondition, "AI provider models require base URL and API key")
	}
	if !isSupportedAIPolishProtocol(cfg.APIProtocol) {
		return nil, status.Error(codes.InvalidArgument, "unsupported AI API protocol")
	}

	lister, ok := ai.NewOpenAICompatiblePolisher(cfg.toRuntimeConfig(server.config)).(ai.ModelLister)
	if !ok {
		return nil, status.Error(codes.Internal, "AI provider does not support listing models")
	}

	models, err := lister.ListModels(ctx)
	if err != nil {
		return nil, mapAIModelsError(err)
	}

	resp := &pb.ListAIModelsResponse{Models: make([]string, 0, len(models))}
	for _, model := range models {
		if id := strings.TrimSpace(model.ID); id != "" {
			resp.Models = append(resp.Models, id)
		}
	}
	return resp, nil
}

func mapAIModelsError(err error) error {
	switch {
	case errors.Is(err, ai.ErrDisabled):
		return status.Error(codes.FailedPrecondition, "AI provider models require base URL and API key")
	case errors.Is(err, ai.ErrProviderFailure):
		return status.Error(codes.Unavailable, "AI provider unavailable")
	case errors.Is(err, ai.ErrMalformedResponse):
		return status.Error(codes.Internal, "AI provider returned an invalid models response")
	default:
		return status.Error(codes.Internal, "failed to list AI models")
	}
}
