package gapi

import (
	"context"
	"errors"
	"testing"
	"time"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	"github.com/MonitorAllen/nostalgia/internal/ai"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeTextPolisher struct {
	response ai.PolishResponse
	err      error
	request  ai.PolishRequest
}

func (polisher *fakeTextPolisher) Polish(ctx context.Context, req ai.PolishRequest) (ai.PolishResponse, error) {
	polisher.request = req
	if polisher.err != nil {
		return ai.PolishResponse{}, polisher.err
	}
	return polisher.response, nil
}

func newPolishTextTestServer(t *testing.T, polisher ai.TextPolisher) *Server {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	server.textPolisher = polisher
	return server
}

func TestPolishTextRequiresAdmin(t *testing.T) {
	server := newPolishTextTestServer(t, &fakeTextPolisher{})
	ctx := newContextWithUserBearerToken(t, server.tokenMaker, util.RandUserID(), "visitor", util.Visitor, time.Minute)

	_, err := server.PolishText(ctx, &pb.PolishTextRequest{
		Mode:   ai.ModeImprove,
		Target: ai.TargetContentSelection,
		Text:   "hello",
	})

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, st.Code())
}

func TestPolishTextRejectsInvalidRequest(t *testing.T) {
	server := newPolishTextTestServer(t, &fakeTextPolisher{})
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	_, err := server.PolishText(ctx, &pb.PolishTextRequest{
		Mode:   "unknown",
		Target: ai.TargetContentSelection,
		Text:   "hello",
	})

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestPolishTextDisabledConfig(t *testing.T) {
	server := newPolishTextTestServer(t, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	_, err := server.PolishText(ctx, &pb.PolishTextRequest{
		Mode:   ai.ModeImprove,
		Target: ai.TargetContentSelection,
		Text:   "hello",
	})

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
	require.Contains(t, st.Message(), "AI 润色尚未配置")
}

func TestPolishTextSuccess(t *testing.T) {
	polisher := &fakeTextPolisher{
		response: ai.PolishResponse{
			Suggestions: []ai.Suggestion{
				{Content: "更自然的表达", Reason: "语气更顺"},
			},
			Mode:   ai.ModeImprove,
			Target: ai.TargetContentSelection,
			Model:  "writer-model",
		},
	}
	server := newPolishTextTestServer(t, polisher)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	resp, err := server.PolishText(ctx, &pb.PolishTextRequest{
		Mode:           ai.ModeImprove,
		Target:         ai.TargetContentSelection,
		Text:           "原始表达",
		ArticleId:      "article-id",
		ArticleTitle:   "文章标题",
		ArticleSummary: "文章摘要",
		ArticleExcerpt: "正文摘录",
		Locale:         "zh-CN",
	})

	require.NoError(t, err)
	require.Equal(t, ai.ModeImprove, polisher.request.Mode)
	require.Equal(t, ai.TargetContentSelection, polisher.request.Target)
	require.Equal(t, "原始表达", polisher.request.Text)
	require.Equal(t, "文章标题", polisher.request.ArticleTitle)
	require.Len(t, resp.GetSuggestions(), 1)
	require.Equal(t, "更自然的表达", resp.GetSuggestions()[0].GetContent())
	require.Equal(t, "语气更顺", resp.GetSuggestions()[0].GetReason())
	require.Equal(t, "writer-model", resp.GetModel())
}

func TestPolishTextMapsProviderFailureWithoutLeakingSecret(t *testing.T) {
	server := newPolishTextTestServer(t, &fakeTextPolisher{
		err: errors.Join(ai.ErrProviderFailure, errors.New("secret-key")),
	})
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	_, err := server.PolishText(ctx, &pb.PolishTextRequest{
		Mode:   ai.ModeImprove,
		Target: ai.TargetContentSelection,
		Text:   "hello",
	})

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unavailable, st.Code())
	require.NotContains(t, st.Message(), "secret-key")
}
