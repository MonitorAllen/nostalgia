package db

import (
	"context"
	"testing"
	"time"

	"github.com/MonitorAllen/nostalgia/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomAutomationArticleRequest(t *testing.T, status string) AutomationArticleRequest {
	t.Helper()

	arg := CreateAutomationArticleRequestParams{
		IdempotencyKey:  "draft-" + uuid.NewString(),
		RequestHash:     util.RandomString(64),
		KeyID:           "codex-daily-writer",
		Status:          status,
		Title:           "Automation draft " + util.RandomString(6),
		SourceTopic:     "Go cache",
		SourcePrompt:    "Write a practical article about cache invalidation.",
		GenerationModel: "codex-automation",
		ErrorMessage:    "",
		ClientIp:        "127.0.0.1",
		UserAgent:       "codex-automation-test",
	}

	request, err := testStore.CreateAutomationArticleRequest(context.Background(), arg)
	require.NoError(t, err)
	require.NotZero(t, request.ID)
	require.Equal(t, arg.IdempotencyKey, request.IdempotencyKey)
	require.Equal(t, arg.RequestHash, request.RequestHash)
	require.Equal(t, arg.KeyID, request.KeyID)
	require.Equal(t, arg.Status, request.Status)
	require.Equal(t, arg.Title, request.Title)
	require.Equal(t, arg.SourceTopic, request.SourceTopic)
	require.Equal(t, arg.SourcePrompt, request.SourcePrompt)
	require.Equal(t, arg.GenerationModel, request.GenerationModel)
	require.Equal(t, arg.ClientIp, request.ClientIp)
	require.Equal(t, arg.UserAgent, request.UserAgent)
	require.False(t, request.ArticleID.Valid)
	require.NotZero(t, request.CreatedAt)
	require.True(t, request.UpdatedAt.IsZero())

	return request
}

func createRandomAdminUser(t *testing.T) User {
	t.Helper()

	hashPassword, err := util.HashPassword(util.RandomString(8))
	require.NoError(t, err)

	user, err := testStore.CreateUserWithRole(context.Background(), CreateUserWithRoleParams{
		ID:              util.RandUserID(),
		Username:        util.RandomOwner(),
		HashedPassword:  hashPassword,
		FullName:        util.RandomOwner(),
		Email:           util.RandomEmail(),
		IsEmailVerified: true,
		Role:            "admin",
	})
	require.NoError(t, err)
	require.Equal(t, "admin", user.Role)

	return user
}

func TestCreateAutomationArticleRequest(t *testing.T) {
	createRandomAutomationArticleRequest(t, "received")
}

func TestGetAutomationArticleRequestByIdempotencyKey(t *testing.T) {
	request := createRandomAutomationArticleRequest(t, "received")

	got, err := testStore.GetAutomationArticleRequestByIdempotencyKey(context.Background(), request.IdempotencyKey)
	require.NoError(t, err)
	require.Equal(t, request.ID, got.ID)
	require.Equal(t, request.RequestHash, got.RequestHash)
	require.Equal(t, request.Status, got.Status)
}

func TestCreateAutomationArticleTx(t *testing.T) {
	owner := createRandomAdminUser(t)
	category := createRandomCategory(t)
	articleID := uuid.New()
	slug := "automation-" + util.RandomString(8)

	result, err := testStore.CreateAutomationArticleTx(context.Background(), CreateAutomationArticleTxParams{
		Request: CreateAutomationArticleRequestParams{
			IdempotencyKey:  "draft-" + uuid.NewString(),
			RequestHash:     util.RandomString(64),
			KeyID:           "codex-daily-writer",
			Status:          "received",
			Title:           "Automation draft " + util.RandomString(6),
			SourceTopic:     "Go cache",
			SourcePrompt:    "Write a practical article about cache invalidation.",
			GenerationModel: "codex-automation",
			ClientIp:        "127.0.0.1",
			UserAgent:       "codex-automation-test",
		},
		Article: CreateAutomationArticleDraftParams{
			ID:            articleID,
			Title:         "Automation draft title",
			Summary:       "Automation draft summary",
			Content:       "<p>Automation draft content</p>",
			Owner:         owner.ID,
			CategoryID:    category.ID,
			Cover:         "https://example.com/cover.jpg",
			Slug:          pgtype.Text{String: slug, Valid: true},
			CheckOutdated: true,
			ReadTime:      "3 min",
		},
	})
	require.NoError(t, err)
	require.Equal(t, articleID, result.Article.ID)
	require.False(t, result.Article.IsPublish)
	require.True(t, result.Article.CreatedByAutomation)
	require.Equal(t, "pending_review", result.Article.AutomationStatus)
	require.True(t, result.Article.AutomationRequestID.Valid)
	require.Equal(t, result.Request.ID, result.Article.AutomationRequestID.Int64)
	require.Equal(t, "created", result.Request.Status)
	require.True(t, result.Request.ArticleID.Valid)
	require.Equal(t, articleID, uuid.UUID(result.Request.ArticleID.Bytes))
	require.NotZero(t, result.Request.UpdatedAt)
}

func TestCountAutomationDraftsToday(t *testing.T) {
	before, err := testStore.CountAutomationDraftsToday(context.Background())
	require.NoError(t, err)

	owner := createRandomAdminUser(t)
	category := createRandomCategory(t)
	_, err = testStore.CreateAutomationArticleTx(context.Background(), CreateAutomationArticleTxParams{
		Request: CreateAutomationArticleRequestParams{
			IdempotencyKey:  "draft-" + uuid.NewString(),
			RequestHash:     util.RandomString(64),
			KeyID:           "codex-daily-writer",
			Status:          "received",
			Title:           "Automation draft " + util.RandomString(6),
			SourceTopic:     "Go cache",
			SourcePrompt:    "Write a practical article about cache invalidation.",
			GenerationModel: "codex-automation",
			ClientIp:        "127.0.0.1",
			UserAgent:       "codex-automation-test",
		},
		Article: CreateAutomationArticleDraftParams{
			ID:            uuid.New(),
			Title:         "Automation draft title",
			Summary:       "Automation draft summary",
			Content:       "<p>Automation draft content</p>",
			Owner:         owner.ID,
			CategoryID:    category.ID,
			Cover:         "",
			Slug:          pgtype.Text{String: "automation-" + util.RandomString(8), Valid: true},
			CheckOutdated: true,
			ReadTime:      "3 min",
		},
	})
	require.NoError(t, err)

	after, err := testStore.CountAutomationDraftsToday(context.Background())
	require.NoError(t, err)
	require.Equal(t, before+1, after)
}

func TestUpdateAutomationArticleMarksPublished(t *testing.T) {
	owner := createRandomAdminUser(t)
	category := createRandomCategory(t)
	result, err := testStore.CreateAutomationArticleTx(context.Background(), CreateAutomationArticleTxParams{
		Request: CreateAutomationArticleRequestParams{
			IdempotencyKey:  "draft-" + uuid.NewString(),
			RequestHash:     util.RandomString(64),
			KeyID:           "codex-daily-writer",
			Status:          "received",
			Title:           "Automation draft " + util.RandomString(6),
			SourceTopic:     "Go cache",
			SourcePrompt:    "Write a practical article about cache invalidation.",
			GenerationModel: "codex-automation",
			ClientIp:        "127.0.0.1",
			UserAgent:       "codex-automation-test",
		},
		Article: CreateAutomationArticleDraftParams{
			ID:            uuid.New(),
			Title:         "Automation draft title",
			Summary:       "Automation draft summary",
			Content:       "<p>Automation draft content</p>",
			Owner:         owner.ID,
			CategoryID:    category.ID,
			Cover:         "",
			Slug:          pgtype.Text{String: "automation-" + util.RandomString(8), Valid: true},
			CheckOutdated: true,
			ReadTime:      "3 min",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "pending_review", result.Article.AutomationStatus)

	updatedArticle, err := testStore.UpdateArticle(context.Background(), UpdateArticleParams{
		ID: result.Article.ID,
		IsPublish: pgtype.Bool{
			Bool:  true,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	})
	require.NoError(t, err)
	require.True(t, updatedArticle.IsPublish)
	require.Equal(t, "published", updatedArticle.AutomationStatus)
}

func TestGetFirstAdminUser(t *testing.T) {
	first := createRandomAdminUser(t)
	time.Sleep(10 * time.Millisecond)
	second := createRandomAdminUser(t)

	got, err := testStore.GetFirstAdminUser(context.Background())
	require.NoError(t, err)
	require.Equal(t, "admin", got.Role)
	require.False(t, got.CreatedAt.After(first.CreatedAt))
	require.False(t, got.CreatedAt.After(second.CreatedAt))
}
