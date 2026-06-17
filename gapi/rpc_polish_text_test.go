package gapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/internal/ai"
	"github.com/MonitorAllen/nostalgia/internal/secrets"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
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
	server := newTestServer(t, nil, nil, nil)
	server.textPolisher = polisher
	return server
}

func encryptTestAIKey(t *testing.T, server *Server, apiKey string) string {
	ciphertext, err := secrets.EncryptString(apiKey, server.config.TokenSymmetricKey, "nostalgia:ai-polish-api-key")
	require.NoError(t, err)
	return ciphertext
}

func testAIConfigRow(t *testing.T, server *Server, baseURL string, apiKey string) db.AiProviderConfig {
	return db.AiProviderConfig{
		Purpose:          "ai_polish",
		Provider:         "openai_compatible",
		ApiProtocol:      ai.APIProtocolChatCompletions,
		BaseUrl:          baseURL,
		Model:            "writer-model",
		ApiKeyCiphertext: encryptTestAIKey(t, server, apiKey),
		TimeoutMs:        int32((2 * time.Second) / time.Millisecond),
		MaxInputChars:    6000,
		MaxContextChars:  4000,
		MaxSuggestions:   3,
		Enabled:          true,
	}
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
		RichText:       "<h2>原始表达</h2><ul><li>保留结构</li></ul>",
		InputFormat:    "html",
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
	require.Equal(t, "<h2>原始表达</h2><ul><li>保留结构</li></ul>", polisher.request.RichText)
	require.Equal(t, "html", polisher.request.InputFormat)
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

func TestGetAIConfigMasksProviderSecret(t *testing.T) {
	server := newPolishTextTestServer(t, nil)
	server.config.AIPolishProvider = "openai_compatible"
	server.config.AIPolishBaseURL = "https://ai.example.com/v1"
	server.config.AIPolishAPIKey = "secret-key"
	server.config.AIPolishModel = "writer-model"
	server.config.AIPolishTimeout = 45 * time.Second
	server.config.AIPolishMaxInputChars = 7000
	server.config.AIPolishMaxContextChars = 5000
	server.config.AIPolishMaxSuggestions = 2
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	resp, err := server.GetAIConfig(ctx, &pb.GetAIConfigRequest{})

	require.NoError(t, err)
	require.Equal(t, "openai_compatible", resp.GetProvider())
	require.Equal(t, "https://ai.example.com/v1", resp.GetBaseUrl())
	require.Equal(t, "writer-model", resp.GetModel())
	require.True(t, resp.GetApiKeyConfigured())
	require.True(t, resp.GetEnabled())
	require.Equal(t, "45s", resp.GetTimeout())
	require.Equal(t, int32(7000), resp.GetMaxInputChars())
	require.Equal(t, int32(5000), resp.GetMaxContextChars())
	require.Equal(t, int32(2), resp.GetMaxSuggestions())
	require.NotContains(t, resp.String(), "secret-key")
}

func TestGetAIConfigUsesDatabaseConfigAndMasksSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	row := testAIConfigRow(t, server, "https://db-ai.example.com/v1", "database-secret")
	store.EXPECT().
		GetAIProviderConfig(gomock.Any(), "ai_polish").
		Return(row, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	resp, err := server.GetAIConfig(ctx, &pb.GetAIConfigRequest{})

	require.NoError(t, err)
	require.Equal(t, "openai_compatible", resp.GetProvider())
	require.Equal(t, "https://db-ai.example.com/v1", resp.GetBaseUrl())
	require.Equal(t, "writer-model", resp.GetModel())
	require.True(t, resp.GetApiKeyConfigured())
	require.True(t, resp.GetEnabled())
	require.Equal(t, "2s", resp.GetTimeout())
	require.Equal(t, "database", resp.GetSource())
	require.NotContains(t, resp.String(), "database-secret")
}

func TestGetAIConfigTreatsUndecryptableDatabaseSecretAsUnconfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	store.EXPECT().
		GetAIProviderConfig(gomock.Any(), "ai_polish").
		Return(db.AiProviderConfig{
			Purpose:          "ai_polish",
			Provider:         "openai_compatible",
			BaseUrl:          "https://db-ai.example.com/v1",
			Model:            "writer-model",
			ApiKeyCiphertext: "v1:invalid:invalid",
			TimeoutMs:        int32((2 * time.Second) / time.Millisecond),
			MaxInputChars:    6000,
			MaxContextChars:  4000,
			MaxSuggestions:   3,
			Enabled:          true,
		}, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	resp, err := server.GetAIConfig(ctx, &pb.GetAIConfigRequest{})

	require.NoError(t, err)
	require.False(t, resp.GetApiKeyConfigured())
	require.False(t, resp.GetEnabled())
	require.Equal(t, "database", resp.GetSource())
}

func TestUpdateAIConfigStoresEncryptedSecretAndMasksResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	store.EXPECT().
		GetAIProviderConfig(gomock.Any(), "ai_polish").
		Return(db.AiProviderConfig{}, pgx.ErrNoRows)
	store.EXPECT().
		UpsertAIProviderConfig(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, arg db.UpsertAIProviderConfigParams) (db.AiProviderConfig, error) {
			require.Equal(t, "ai_polish", arg.Purpose)
			require.Equal(t, "DeepSeek", arg.Provider)
			require.Equal(t, ai.APIProtocolResponses, arg.ApiProtocol)
			require.Equal(t, "https://ai.example.com/v1", arg.BaseUrl)
			require.Equal(t, "writer-model", arg.Model)
			require.Equal(t, int32(45000), arg.TimeoutMs)
			require.Equal(t, int32(7000), arg.MaxInputChars)
			require.Equal(t, int32(5000), arg.MaxContextChars)
			require.Equal(t, int32(2), arg.MaxSuggestions)
			require.True(t, arg.Enabled)
			require.True(t, arg.UpdatedBy.Valid)
			require.NotEmpty(t, arg.ApiKeyCiphertext)
			require.NotContains(t, arg.ApiKeyCiphertext, "new-secret")

			plaintext, err := secrets.DecryptString(arg.ApiKeyCiphertext, server.config.TokenSymmetricKey, "nostalgia:ai-polish-api-key")
			require.NoError(t, err)
			require.Equal(t, "new-secret", plaintext)

			return db.AiProviderConfig{
				Purpose:          arg.Purpose,
				Provider:         arg.Provider,
				ApiProtocol:      arg.ApiProtocol,
				BaseUrl:          arg.BaseUrl,
				Model:            arg.Model,
				ApiKeyCiphertext: arg.ApiKeyCiphertext,
				TimeoutMs:        arg.TimeoutMs,
				MaxInputChars:    arg.MaxInputChars,
				MaxContextChars:  arg.MaxContextChars,
				MaxSuggestions:   arg.MaxSuggestions,
				Enabled:          arg.Enabled,
				UpdatedBy:        arg.UpdatedBy,
			}, nil
		})
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	resp, err := server.UpdateAIConfig(ctx, &pb.UpdateAIConfigRequest{
		Provider:        "DeepSeek",
		ApiProtocol:     ai.APIProtocolResponses,
		BaseUrl:         "https://ai.example.com/v1",
		Model:           "writer-model",
		ApiKey:          "new-secret",
		Timeout:         "45s",
		MaxInputChars:   7000,
		MaxContextChars: 5000,
		MaxSuggestions:  2,
		Enabled:         true,
	})

	require.NoError(t, err)
	require.Equal(t, "DeepSeek", resp.GetProvider())
	require.Equal(t, ai.APIProtocolResponses, resp.GetApiProtocol())
	require.Equal(t, "database", resp.GetSource())
	require.True(t, resp.GetApiKeyConfigured())
	require.True(t, resp.GetEnabled())
	require.NotContains(t, resp.String(), "new-secret")
}

func TestUpdateAIConfigStoresPromptTemplates(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	store.EXPECT().
		GetAIProviderConfig(gomock.Any(), "ai_polish").
		Return(db.AiProviderConfig{}, pgx.ErrNoRows)
	store.EXPECT().
		UpsertAIProviderConfig(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, arg db.UpsertAIProviderConfigParams) (db.AiProviderConfig, error) {
			require.JSONEq(t, `{"improve":"custom {{text}}"}`, string(arg.PromptTemplates))
			return db.AiProviderConfig{
				Purpose:          arg.Purpose,
				Provider:         arg.Provider,
				ApiProtocol:      arg.ApiProtocol,
				BaseUrl:          arg.BaseUrl,
				Model:            arg.Model,
				ApiKeyCiphertext: arg.ApiKeyCiphertext,
				TimeoutMs:        arg.TimeoutMs,
				MaxInputChars:    arg.MaxInputChars,
				MaxContextChars:  arg.MaxContextChars,
				MaxSuggestions:   arg.MaxSuggestions,
				PromptTemplates:  arg.PromptTemplates,
				Enabled:          arg.Enabled,
				UpdatedBy:        arg.UpdatedBy,
			}, nil
		})

	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)
	resp, err := server.UpdateAIConfig(ctx, &pb.UpdateAIConfigRequest{
		Provider:        "openai",
		ApiProtocol:     ai.APIProtocolChatCompletions,
		BaseUrl:         "https://ai.example.com/v1",
		Model:           "writer-model",
		ApiKey:          "new-secret",
		Timeout:         "30s",
		MaxInputChars:   6000,
		MaxContextChars: 4000,
		MaxSuggestions:  3,
		Enabled:         true,
		PromptTemplates: map[string]string{ai.ModeImprove: "custom {{text}}"},
	})

	require.NoError(t, err)
	require.Equal(t, "custom {{text}}", resp.GetPromptTemplates()[ai.ModeImprove])
	require.NotEmpty(t, resp.GetDefaultPromptTemplates()[ai.ModeImprove])
}

func TestUpdateAIConfigRejectsEnabledConfigWithoutAPIKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	store.EXPECT().
		GetAIProviderConfig(gomock.Any(), "ai_polish").
		Return(db.AiProviderConfig{}, pgx.ErrNoRows)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	_, err := server.UpdateAIConfig(ctx, &pb.UpdateAIConfigRequest{
		Provider:        "openai_compatible",
		BaseUrl:         "https://ai.example.com/v1",
		Model:           "writer-model",
		Timeout:         "30s",
		MaxInputChars:   6000,
		MaxContextChars: 4000,
		MaxSuggestions:  3,
		Enabled:         true,
	})

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestListAIModelsUsesProviderModelsEndpointWithoutSaving(t *testing.T) {
	var gotPath string
	var gotAuth string
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"writer-model"},{"id":"fast-model"}]}`))
	}))
	defer provider.Close()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	row := testAIConfigRow(t, server, provider.URL+"/v1", "database-secret")
	store.EXPECT().
		GetAIProviderConfig(gomock.Any(), "ai_polish").
		Return(row, nil)
	store.EXPECT().
		UpsertAIProviderConfig(gomock.Any(), gomock.Any()).
		Times(0)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	resp, err := server.ListAIModels(ctx, &pb.ListAIModelsRequest{
		Provider:    "DeepSeek",
		ApiProtocol: ai.APIProtocolChatCompletions,
		BaseUrl:     provider.URL + "/v1",
	})

	require.NoError(t, err)
	require.Equal(t, "/v1/models", gotPath)
	require.Equal(t, "Bearer database-secret", gotAuth)
	require.Equal(t, []string{"writer-model", "fast-model"}, resp.GetModels())
}

func TestListAIModelsReturnsUnimplementedForAnthropic(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"claude-test"}]}`))
	}))
	defer provider.Close()

	server := newTestServer(t, nil, nil, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	_, err := server.ListAIModels(ctx, &pb.ListAIModelsRequest{
		Provider:    ai.ProviderAnthropic,
		ApiProtocol: ai.APIProtocolMessages,
		BaseUrl:     provider.URL,
		ApiKey:      "secret-key",
	})

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unimplemented, st.Code())
}

func TestPolishTextUsesDatabaseConfig(t *testing.T) {
	var gotAuth string
	var gotModel string
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")

		var payload struct {
			Model string `json:"model"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		gotModel = payload.Model

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"suggestions\":[{\"content\":\"新表达\",\"reason\":\"更清楚\"}]}"}}]}`))
	}))
	defer provider.Close()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	row := testAIConfigRow(t, server, provider.URL, "database-secret")
	store.EXPECT().
		GetAIProviderConfig(gomock.Any(), "ai_polish").
		Return(row, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	resp, err := server.PolishText(ctx, &pb.PolishTextRequest{
		Mode:   ai.ModeImprove,
		Target: ai.TargetContentSelection,
		Text:   "原始表达",
	})

	require.NoError(t, err)
	require.Equal(t, "Bearer database-secret", gotAuth)
	require.Equal(t, "writer-model", gotModel)
	require.Len(t, resp.GetSuggestions(), 1)
	require.Equal(t, "新表达", resp.GetSuggestions()[0].GetContent())
}

func TestGetAIConfigRequiresAdmin(t *testing.T) {
	server := newPolishTextTestServer(t, nil)
	ctx := newContextWithUserBearerToken(t, server.tokenMaker, util.RandUserID(), "visitor", util.Visitor, time.Minute)

	_, err := server.GetAIConfig(ctx, &pb.GetAIConfigRequest{})

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, st.Code())
}
