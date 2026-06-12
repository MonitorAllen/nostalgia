package gapi

import (
	"context"
	"errors"

	"github.com/MonitorAllen/nostalgia/internal/ai"
	"github.com/MonitorAllen/nostalgia/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) PolishText(ctx context.Context, req *pb.PolishTextRequest) (*pb.PolishTextResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	polishReq := ai.PolishRequest{
		Mode:           req.GetMode(),
		Target:         req.GetTarget(),
		Text:           req.GetText(),
		ArticleID:      req.GetArticleId(),
		ArticleTitle:   req.GetArticleTitle(),
		ArticleSummary: req.GetArticleSummary(),
		ArticleExcerpt: req.GetArticleExcerpt(),
		Locale:         req.GetLocale(),
	}

	cfg := server.runtimeAIPolishConfig()
	polisher := server.textPolisher
	if polisher == nil {
		var resolveErr error
		cfg, resolveErr = server.resolveAIPolishConfig(ctx)
		if resolveErr != nil {
			return nil, status.Error(codes.Internal, "failed to load AI config")
		}
	}

	if err := ai.ValidateRequest(polishReq, cfg.MaxInputChars); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid AI polish request")
	}

	if polisher == nil {
		if !cfg.usable() {
			return nil, status.Error(codes.FailedPrecondition, "AI 润色尚未配置")
		}
		polisher = ai.NewOpenAICompatiblePolisher(cfg.toRuntimeConfig(server.config))
	}
	if polisher == nil {
		return nil, status.Error(codes.FailedPrecondition, "AI 润色尚未配置")
	}

	result, err := polisher.Polish(ctx, polishReq)
	if err != nil {
		return nil, mapAIPolishError(err)
	}

	suggestions := make([]*pb.PolishSuggestion, 0, len(result.Suggestions))
	for _, suggestion := range result.Suggestions {
		suggestions = append(suggestions, &pb.PolishSuggestion{
			Content: suggestion.Content,
			Reason:  suggestion.Reason,
		})
	}

	return &pb.PolishTextResponse{
		Suggestions: suggestions,
		Mode:        result.Mode,
		Target:      result.Target,
		Model:       result.Model,
	}, nil
}

func mapAIPolishError(err error) error {
	switch {
	case errors.Is(err, ai.ErrDisabled):
		return status.Error(codes.FailedPrecondition, "AI 润色尚未配置")
	case errors.Is(err, ai.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid AI polish request")
	case errors.Is(err, ai.ErrProviderFailure):
		return status.Error(codes.Unavailable, "AI provider unavailable")
	case errors.Is(err, ai.ErrMalformedResponse):
		return status.Error(codes.Internal, "AI provider returned an invalid response")
	default:
		return status.Error(codes.Internal, "AI polish failed")
	}
}
