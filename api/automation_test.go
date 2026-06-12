package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/internal/automation"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/MonitorAllen/nostalgia/worker"
	mockwk "github.com/MonitorAllen/nostalgia/worker/mock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

type automationDraftTestBody struct {
	Title           string `json:"title"`
	Summary         string `json:"summary"`
	Content         string `json:"content"`
	CategoryID      int64  `json:"category_id"`
	Slug            string `json:"slug,omitempty"`
	Cover           string `json:"cover,omitempty"`
	CheckOutdated   bool   `json:"check_outdated"`
	SourceTopic     string `json:"source_topic,omitempty"`
	SourcePrompt    string `json:"source_prompt,omitempty"`
	GenerationModel string `json:"generation_model,omitempty"`
}

func TestCreateAutomationArticleDraftOK(t *testing.T) {
	now := time.Now()
	body := defaultAutomationDraftTestBody()
	rawBody := mustMarshalAutomationDraftBody(t, body)
	idempotencyKey := "2026-06-12-codex-go-cache"
	requestHash := automation.SHA256Hex(rawBody)
	articleID := uuid.New()
	owner := db.User{ID: util.RandUserID(), Role: util.Admin}
	category := db.Category{ID: body.CategoryID, Name: "Go"}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	taskDistributor := mockwk.NewMockTaskDistributor(ctrl)

	store.EXPECT().
		GetAutomationArticleRequestByIdempotencyKey(gomock.Any(), gomock.Eq(idempotencyKey)).
		Times(1).
		Return(db.AutomationArticleRequest{}, db.ErrRecordNotFound)
	store.EXPECT().CountAutomationDraftsToday(gomock.Any()).Times(1).Return(int64(0), nil)
	store.EXPECT().GetCategory(gomock.Any(), gomock.Eq(body.CategoryID)).Times(1).Return(category, nil)
	store.EXPECT().
		GetArticleBySlug(gomock.Any(), gomock.Eq(pgtype.Text{String: body.Slug, Valid: true})).
		Times(1).
		Return(db.GetArticleBySlugRow{}, db.ErrRecordNotFound)
	store.EXPECT().GetFirstAdminUser(gomock.Any()).Times(1).Return(owner, nil)
	store.EXPECT().
		CreateAutomationArticleTx(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(ctx context.Context, arg db.CreateAutomationArticleTxParams) (db.CreateAutomationArticleTxResult, error) {
			require.Equal(t, idempotencyKey, arg.Request.IdempotencyKey)
			require.Equal(t, requestHash, arg.Request.RequestHash)
			require.Equal(t, "codex-daily-writer", arg.Request.KeyID)
			require.Equal(t, "received", arg.Request.Status)
			require.Equal(t, body.Title, arg.Request.Title)
			require.Equal(t, body.SourceTopic, arg.Request.SourceTopic)
			require.Equal(t, body.SourcePrompt, arg.Request.SourcePrompt)
			require.Equal(t, body.GenerationModel, arg.Request.GenerationModel)
			require.Equal(t, body.Title, arg.Article.Title)
			require.Equal(t, body.Summary, arg.Article.Summary)
			require.Equal(t, body.Content, arg.Article.Content)
			require.Equal(t, owner.ID, arg.Article.Owner)
			require.Equal(t, category.ID, arg.Article.CategoryID)
			require.Equal(t, body.Slug, arg.Article.Slug.String)
			require.True(t, arg.Article.Slug.Valid)
			require.Equal(t, "1 分钟", arg.Article.ReadTime)

			return db.CreateAutomationArticleTxResult{
				Request: db.AutomationArticleRequest{
					ID:             99,
					IdempotencyKey: idempotencyKey,
					RequestHash:    requestHash,
					Status:         "created",
					ArticleID:      pgtype.UUID{Bytes: articleID, Valid: true},
					Title:          body.Title,
				},
				Article: db.Article{
					ID:                  articleID,
					Title:               body.Title,
					IsPublish:           false,
					CreatedByAutomation: true,
					AutomationStatus:    "pending_review",
				},
			}, nil
		})
	taskDistributor.EXPECT().
		DistributeTaskNotifyAutomationDraft(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(ctx context.Context, payload *worker.PayloadNotifyAutomationDraft, opts ...any) error {
			require.Equal(t, "success", payload.Kind)
			require.Equal(t, articleID, payload.ArticleID)
			require.Equal(t, body.Title, payload.Title)
			require.Equal(t, category.Name, payload.CategoryName)
			require.Equal(t, "https://example.com/backend/articles/"+articleID.String(), payload.ReviewURL)
			require.Equal(t, idempotencyKey, payload.IdempotencyKey)
			require.Equal(t, body.GenerationModel, payload.GenerationModel)
			require.Equal(t, "owner@example.com", payload.NotifyEmail)
			return nil
		})

	server := newAutomationTestServer(t, store, taskDistributor)
	recorder := httptest.NewRecorder()
	request := newSignedAutomationDraftRequest(t, rawBody, now, idempotencyKey, "codex-daily-writer", "secret")

	server.router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusCreated, recorder.Code)
	requireAutomationDraftResponse(t, recorder.Body, articleID, "created")
}

func TestCreateAutomationArticleDraftAuthFailuresDoNotNotify(t *testing.T) {
	now := time.Now()
	body := defaultAutomationDraftTestBody()
	rawBody := mustMarshalAutomationDraftBody(t, body)

	testCases := []struct {
		name    string
		request func(t *testing.T) *http.Request
	}{
		{
			name: "MissingSignature",
			request: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/api/automation/articles/drafts", bytes.NewReader(rawBody))
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
		},
		{
			name: "InvalidSignature",
			request: func(t *testing.T) *http.Request {
				req := newSignedAutomationDraftRequest(t, rawBody, now, "draft-invalid-signature", "codex-daily-writer", "secret")
				req.Header.Set("X-Automation-Signature", "v1=bad")
				return req
			},
		},
		{
			name: "ExpiredTimestamp",
			request: func(t *testing.T) *http.Request {
				return newSignedAutomationDraftRequest(t, rawBody, now.Add(-10*time.Minute), "draft-expired", "codex-daily-writer", "secret")
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			taskDistributor := mockwk.NewMockTaskDistributor(ctrl)
			server := newAutomationTestServer(t, store, taskDistributor)
			server.config.AutomationSignatureTTL = 5 * time.Minute
			recorder := httptest.NewRecorder()

			server.router.ServeHTTP(recorder, tc.request(t))

			require.Equal(t, http.StatusUnauthorized, recorder.Code)
			require.NotContains(t, recorder.Body.String(), "secret")
		})
	}
}

func TestCreateAutomationArticleDraftDisabledConfigReturnsNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	server := newAutomationTestServer(t, store, nil)
	server.config.AutomationHMACSecret = ""
	recorder := httptest.NewRecorder()
	body := mustMarshalAutomationDraftBody(t, defaultAutomationDraftTestBody())
	request := newSignedAutomationDraftRequest(t, body, time.Now(), "draft-disabled", "codex-daily-writer", "secret")

	server.router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestCreateAutomationArticleDraftIdempotentReplay(t *testing.T) {
	now := time.Now()
	body := defaultAutomationDraftTestBody()
	rawBody := mustMarshalAutomationDraftBody(t, body)
	idempotencyKey := "draft-replay"
	articleID := uuid.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().
		GetAutomationArticleRequestByIdempotencyKey(gomock.Any(), gomock.Eq(idempotencyKey)).
		Times(1).
		Return(db.AutomationArticleRequest{
			IdempotencyKey: idempotencyKey,
			RequestHash:    automation.SHA256Hex(rawBody),
			Status:         "created",
			ArticleID:      pgtype.UUID{Bytes: articleID, Valid: true},
			Title:          body.Title,
		}, nil)

	server := newAutomationTestServer(t, store, nil)
	recorder := httptest.NewRecorder()
	request := newSignedAutomationDraftRequest(t, rawBody, now, idempotencyKey, "codex-daily-writer", "secret")

	server.router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	requireAutomationDraftResponse(t, recorder.Body, articleID, "replayed")
}

func TestCreateAutomationArticleDraftIdempotencyConflict(t *testing.T) {
	now := time.Now()
	body := defaultAutomationDraftTestBody()
	rawBody := mustMarshalAutomationDraftBody(t, body)
	idempotencyKey := "draft-conflict"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().
		GetAutomationArticleRequestByIdempotencyKey(gomock.Any(), gomock.Eq(idempotencyKey)).
		Times(1).
		Return(db.AutomationArticleRequest{
			IdempotencyKey: idempotencyKey,
			RequestHash:    strings.Repeat("a", 64),
			Status:         "created",
		}, nil)

	server := newAutomationTestServer(t, store, nil)
	recorder := httptest.NewRecorder()
	request := newSignedAutomationDraftRequest(t, rawBody, now, idempotencyKey, "codex-daily-writer", "secret")

	server.router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusConflict, recorder.Code)
}

func TestCreateAutomationArticleDraftMissingCategoryRecordsFailure(t *testing.T) {
	now := time.Now()
	body := defaultAutomationDraftTestBody()
	body.CategoryID = 404
	rawBody := mustMarshalAutomationDraftBody(t, body)
	idempotencyKey := "draft-missing-category"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	taskDistributor := mockwk.NewMockTaskDistributor(ctrl)
	expectFailureAuditAndEmail(t, store, taskDistributor, idempotencyKey, body.Title, automation.SHA256Hex(rawBody), "failed_validation")
	store.EXPECT().
		GetAutomationArticleRequestByIdempotencyKey(gomock.Any(), gomock.Eq(idempotencyKey)).
		Times(1).
		Return(db.AutomationArticleRequest{}, db.ErrRecordNotFound)
	store.EXPECT().CountAutomationDraftsToday(gomock.Any()).Times(1).Return(int64(0), nil)
	store.EXPECT().GetCategory(gomock.Any(), gomock.Eq(body.CategoryID)).Times(1).Return(db.Category{}, db.ErrRecordNotFound)

	server := newAutomationTestServer(t, store, taskDistributor)
	recorder := httptest.NewRecorder()
	request := newSignedAutomationDraftRequest(t, rawBody, now, idempotencyKey, "codex-daily-writer", "secret")

	server.router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
}

func TestCreateAutomationArticleDraftSlugConflictRecordsFailure(t *testing.T) {
	now := time.Now()
	body := defaultAutomationDraftTestBody()
	rawBody := mustMarshalAutomationDraftBody(t, body)
	idempotencyKey := "draft-slug-conflict"
	category := db.Category{ID: body.CategoryID, Name: "Go"}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	taskDistributor := mockwk.NewMockTaskDistributor(ctrl)
	expectFailureAuditAndEmail(t, store, taskDistributor, idempotencyKey, body.Title, automation.SHA256Hex(rawBody), "failed_validation")
	store.EXPECT().
		GetAutomationArticleRequestByIdempotencyKey(gomock.Any(), gomock.Eq(idempotencyKey)).
		Times(1).
		Return(db.AutomationArticleRequest{}, db.ErrRecordNotFound)
	store.EXPECT().CountAutomationDraftsToday(gomock.Any()).Times(1).Return(int64(0), nil)
	store.EXPECT().GetCategory(gomock.Any(), gomock.Eq(body.CategoryID)).Times(1).Return(category, nil)
	store.EXPECT().
		GetArticleBySlug(gomock.Any(), gomock.Eq(pgtype.Text{String: body.Slug, Valid: true})).
		Times(1).
		Return(db.GetArticleBySlugRow{ID: uuid.New(), Slug: pgtype.Text{String: body.Slug, Valid: true}}, nil)

	server := newAutomationTestServer(t, store, taskDistributor)
	recorder := httptest.NewRecorder()
	request := newSignedAutomationDraftRequest(t, rawBody, now, idempotencyKey, "codex-daily-writer", "secret")

	server.router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusConflict, recorder.Code)
}

func TestCreateAutomationArticleDraftDailyLimitRecordsFailure(t *testing.T) {
	now := time.Now()
	body := defaultAutomationDraftTestBody()
	rawBody := mustMarshalAutomationDraftBody(t, body)
	idempotencyKey := "draft-daily-limit"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	taskDistributor := mockwk.NewMockTaskDistributor(ctrl)
	expectFailureAuditAndEmail(t, store, taskDistributor, idempotencyKey, body.Title, automation.SHA256Hex(rawBody), "failed_validation")
	store.EXPECT().
		GetAutomationArticleRequestByIdempotencyKey(gomock.Any(), gomock.Eq(idempotencyKey)).
		Times(1).
		Return(db.AutomationArticleRequest{}, db.ErrRecordNotFound)
	store.EXPECT().CountAutomationDraftsToday(gomock.Any()).Times(1).Return(int64(1), nil)

	server := newAutomationTestServer(t, store, taskDistributor)
	recorder := httptest.NewRecorder()
	request := newSignedAutomationDraftRequest(t, rawBody, now, idempotencyKey, "codex-daily-writer", "secret")

	server.router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusConflict, recorder.Code)
}

func newAutomationTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	config := util.Config{
		TokenSymmetricKey:         util.RandomString(32),
		AccessTokenDuration:       time.Minute,
		Domain:                    "https://example.com",
		AutomationHMACKeyID:       "codex-daily-writer",
		AutomationHMACSecret:      "secret",
		AutomationSignatureTTL:    5 * time.Minute,
		AutomationDailyDraftLimit: 1,
		AutomationNotifyEmail:     "owner@example.com",
		EmailSenderAddress:        "sender@example.com",
		UploadFileSizeLimit:       2 << 20,
	}

	server, err := NewServer(config, store, taskDistributor, nil)
	require.NoError(t, err)
	return server
}

func defaultAutomationDraftTestBody() automationDraftTestBody {
	return automationDraftTestBody{
		Title:           "Redis cache invalidation",
		Summary:         "A practical guide to cache invalidation.",
		Content:         "<p>Hello cache.</p>",
		CategoryID:      1,
		Slug:            "redis-cache-invalidation",
		Cover:           "https://example.com/cover.jpg",
		CheckOutdated:   true,
		SourceTopic:     "Go cache",
		SourcePrompt:    "Write a practical article about cache invalidation.",
		GenerationModel: "codex-automation",
	}
}

func mustMarshalAutomationDraftBody(t *testing.T, body automationDraftTestBody) []byte {
	t.Helper()

	data, err := json.Marshal(body)
	require.NoError(t, err)
	return data
}

func newSignedAutomationDraftRequest(t *testing.T, rawBody []byte, now time.Time, idempotencyKey string, keyID string, secret string) *http.Request {
	t.Helper()

	timestamp := now.Format(time.RFC3339)
	base := automation.SignatureBaseString(
		http.MethodPost,
		"/api/automation/articles/drafts",
		timestamp,
		idempotencyKey,
		automation.SHA256Hex(rawBody),
	)
	signature := "v1=" + automation.Sign(secret, base)

	request, err := http.NewRequest(http.MethodPost, "/api/automation/articles/drafts", bytes.NewReader(rawBody))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Automation-Key-Id", keyID)
	request.Header.Set("X-Automation-Timestamp", timestamp)
	request.Header.Set("X-Automation-Signature", signature)
	request.Header.Set("Idempotency-Key", idempotencyKey)
	return request
}

func expectFailureAuditAndEmail(
	t *testing.T,
	store *mockdb.MockStore,
	taskDistributor *mockwk.MockTaskDistributor,
	idempotencyKey string,
	title string,
	requestHash string,
	status string,
) {
	store.EXPECT().
		CreateAutomationArticleRequest(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(ctx context.Context, arg db.CreateAutomationArticleRequestParams) (db.AutomationArticleRequest, error) {
			require.Equal(t, idempotencyKey, arg.IdempotencyKey)
			require.Equal(t, requestHash, arg.RequestHash)
			require.Equal(t, "codex-daily-writer", arg.KeyID)
			require.Equal(t, status, arg.Status)
			require.Equal(t, title, arg.Title)
			require.NotEmpty(t, arg.ErrorMessage)
			return db.AutomationArticleRequest{
				ID:             101,
				IdempotencyKey: idempotencyKey,
				RequestHash:    requestHash,
				Status:         status,
				Title:          title,
				ErrorMessage:   arg.ErrorMessage,
			}, nil
		})
	taskDistributor.EXPECT().
		DistributeTaskNotifyAutomationDraft(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(ctx context.Context, payload *worker.PayloadNotifyAutomationDraft, opts ...any) error {
			require.Equal(t, "failure", payload.Kind)
			require.Equal(t, title, payload.Title)
			require.Equal(t, idempotencyKey, payload.IdempotencyKey)
			require.NotEmpty(t, payload.ErrorMessage)
			require.Equal(t, "owner@example.com", payload.NotifyEmail)
			return nil
		})
}

func requireAutomationDraftResponse(t *testing.T, body *bytes.Buffer, articleID uuid.UUID, status string) {
	t.Helper()

	var response struct {
		Article struct {
			ID                  uuid.UUID `json:"id"`
			IsPublish           bool      `json:"is_publish"`
			CreatedByAutomation bool      `json:"created_by_automation"`
			AutomationStatus    string    `json:"automation_status"`
		} `json:"article"`
		ReviewURL      string `json:"review_url"`
		IdempotencyKey string `json:"idempotency_key"`
		Status         string `json:"status"`
	}
	err := json.Unmarshal(body.Bytes(), &response)
	require.NoError(t, err)
	require.Equal(t, articleID, response.Article.ID)
	require.False(t, response.Article.IsPublish)
	require.True(t, response.Article.CreatedByAutomation)
	require.Equal(t, "pending_review", response.Article.AutomationStatus)
	require.Equal(t, "https://example.com/backend/articles/"+articleID.String(), response.ReviewURL)
	require.Equal(t, status, response.Status)
}
