package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	automationauth "github.com/MonitorAllen/nostalgia/internal/automation"
	"github.com/MonitorAllen/nostalgia/worker"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

const (
	automationKeyIDHeader     = "X-Automation-Key-Id"
	automationTimestampHeader = "X-Automation-Timestamp"
	automationSignatureHeader = "X-Automation-Signature"
	idempotencyKeyHeader      = "Idempotency-Key"

	automationRequestStatusReceived         = "received"
	automationRequestStatusCreated          = "created"
	automationRequestStatusFailedValidation = "failed_validation"
	automationRequestStatusFailedCreate     = "failed_create"
)

var automationSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type createAutomationArticleDraftRequest struct {
	Title           string `json:"title" binding:"required"`
	Summary         string `json:"summary" binding:"required"`
	Content         string `json:"content" binding:"required"`
	CategoryID      int64  `json:"category_id" binding:"required,min=1"`
	Slug            string `json:"slug" binding:"omitempty,min=5,max=100"`
	Cover           string `json:"cover"`
	CheckOutdated   *bool  `json:"check_outdated"`
	SourceTopic     string `json:"source_topic"`
	SourcePrompt    string `json:"source_prompt"`
	GenerationModel string `json:"generation_model"`
}

type automationDraftArticleResponse struct {
	ID                  uuid.UUID `json:"id"`
	Title               string    `json:"title"`
	IsPublish           bool      `json:"is_publish"`
	CreatedByAutomation bool      `json:"created_by_automation"`
	AutomationStatus    string    `json:"automation_status"`
}

type automationDraftResponse struct {
	Article        automationDraftArticleResponse `json:"article"`
	ReviewURL      string                         `json:"review_url"`
	IdempotencyKey string                         `json:"idempotency_key"`
	Status         string                         `json:"status"`
}

func (server *Server) createAutomationArticleDraft(ctx *gin.Context) {
	if !server.automationDraftAPIEnabled() {
		ctx.JSON(http.StatusNotFound, errorResponse(errors.New("not found")))
		return
	}

	rawBody, err := readAutomationDraftBody(ctx, server.config.UploadFileSizeLimit)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	idempotencyKey := ctx.GetHeader(idempotencyKeyHeader)
	keyID := ctx.GetHeader(automationKeyIDHeader)
	authErr := automationauth.VerifySignature(automationauth.SignatureInput{
		Method:         ctx.Request.Method,
		Path:           ctx.Request.URL.Path,
		Timestamp:      ctx.GetHeader(automationTimestampHeader),
		IdempotencyKey: idempotencyKey,
		Body:           rawBody,
		Now:            time.Now(),
		TTL:            server.config.AutomationSignatureTTL,
		KeyID:          keyID,
		ExpectedKeyID:  server.config.AutomationHMACKeyID,
		Secret:         server.config.AutomationHMACSecret,
		Signature:      ctx.GetHeader(automationSignatureHeader),
	})
	requestHash := automationauth.SHA256Hex(rawBody)
	if authErr != nil {
		log.Warn().
			Err(authErr).
			Str("module", "automation").
			Str("action", "create_article_draft").
			Str("key_id", keyID).
			Str("idempotency_key", idempotencyKey).
			Str("request_hash", requestHash).
			Str("client_ip", ctx.ClientIP()).
			Msg("automation draft authentication failed")
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("invalid automation authentication")))
		return
	}

	ctx.Request.Body = io.NopCloser(bytes.NewReader(rawBody))
	var req createAutomationArticleDraftRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	req.normalize()
	if err := req.validate(); err != nil {
		if recordErr := server.recordAutomationDraftFailure(ctx, req, automationRequestStatusFailedValidation, requestHash, idempotencyKey, keyID, err); recordErr != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(recordErr))
			return
		}
		ctx.JSON(http.StatusUnprocessableEntity, errorResponse(err))
		return
	}

	existingRequest, err := server.store.GetAutomationArticleRequestByIdempotencyKey(ctx, idempotencyKey)
	if err == nil {
		server.handleAutomationDraftReplay(ctx, existingRequest, requestHash, idempotencyKey)
		return
	}
	if !errors.Is(err, db.ErrRecordNotFound) {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	draftsToday, err := server.store.CountAutomationDraftsToday(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if draftsToday >= server.automationDailyDraftLimit() {
		err := fmt.Errorf("daily automation draft limit reached")
		if recordErr := server.recordAutomationDraftFailure(ctx, req, automationRequestStatusFailedValidation, requestHash, idempotencyKey, keyID, err); recordErr != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(recordErr))
			return
		}
		ctx.JSON(http.StatusConflict, errorResponse(err))
		return
	}

	category, err := server.store.GetCategory(ctx, req.CategoryID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, db.ErrRecordNotFound) {
			err = fmt.Errorf("category does not exist")
			statusCode = http.StatusUnprocessableEntity
		}
		if statusCode != http.StatusInternalServerError {
			if recordErr := server.recordAutomationDraftFailure(ctx, req, automationRequestStatusFailedValidation, requestHash, idempotencyKey, keyID, err); recordErr != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(recordErr))
				return
			}
		}
		ctx.JSON(statusCode, errorResponse(err))
		return
	}

	if req.Slug != "" {
		_, err := server.store.GetArticleBySlug(ctx, pgtype.Text{String: req.Slug, Valid: true})
		if err == nil {
			err = fmt.Errorf("slug already exists")
			if recordErr := server.recordAutomationDraftFailure(ctx, req, automationRequestStatusFailedValidation, requestHash, idempotencyKey, keyID, err); recordErr != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(recordErr))
				return
			}
			ctx.JSON(http.StatusConflict, errorResponse(err))
			return
		}
		if !errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	owner, err := server.store.GetFirstAdminUser(ctx)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, db.ErrRecordNotFound) {
			err = fmt.Errorf("admin owner does not exist")
			statusCode = http.StatusUnprocessableEntity
		}
		if statusCode != http.StatusInternalServerError {
			if recordErr := server.recordAutomationDraftFailure(ctx, req, automationRequestStatusFailedValidation, requestHash, idempotencyKey, keyID, err); recordErr != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(recordErr))
				return
			}
		}
		ctx.JSON(statusCode, errorResponse(err))
		return
	}

	articleID := uuid.New()
	result, err := server.store.CreateAutomationArticleTx(ctx, db.CreateAutomationArticleTxParams{
		Request: db.CreateAutomationArticleRequestParams{
			IdempotencyKey:  idempotencyKey,
			RequestHash:     requestHash,
			KeyID:           keyID,
			Status:          automationRequestStatusReceived,
			Title:           req.Title,
			SourceTopic:     req.SourceTopic,
			SourcePrompt:    req.SourcePrompt,
			GenerationModel: req.GenerationModel,
			ErrorMessage:    "",
			ClientIp:        ctx.ClientIP(),
			UserAgent:       ctx.Request.UserAgent(),
		},
		Article: db.CreateAutomationArticleDraftParams{
			ID:            articleID,
			Title:         req.Title,
			Summary:       req.Summary,
			Content:       req.Content,
			Owner:         owner.ID,
			CategoryID:    category.ID,
			Cover:         req.Cover,
			Slug:          pgtype.Text{String: req.Slug, Valid: req.Slug != ""},
			CheckOutdated: req.checkOutdated(),
			ReadTime:      calculateAutomationReadTime(req.Content),
		},
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if db.ErrorCode(err) == db.UniqueViolation {
			statusCode = http.StatusConflict
		}
		if recordErr := server.recordAutomationDraftFailure(ctx, req, automationRequestStatusFailedCreate, requestHash, idempotencyKey, keyID, err); recordErr != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(recordErr))
			return
		}
		ctx.JSON(statusCode, errorResponse(err))
		return
	}

	reviewURL := server.automationReviewURL(result.Article.ID)
	if err := server.enqueueAutomationDraftSuccess(ctx, req, category.Name, result.Article.ID, reviewURL, idempotencyKey); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info().
		Str("module", "automation").
		Str("action", "create_article_draft").
		Str("key_id", keyID).
		Str("idempotency_key", idempotencyKey).
		Str("request_hash", requestHash).
		Str("article_id", result.Article.ID.String()).
		Str("status", automationRequestStatusCreated).
		Str("client_ip", ctx.ClientIP()).
		Msg("automation draft created")

	ctx.JSON(http.StatusCreated, server.newAutomationDraftResponse(result.Article.ID, result.Article.Title, idempotencyKey, "created"))
}

func readAutomationDraftBody(ctx *gin.Context, configuredLimit int64) ([]byte, error) {
	limit := configuredLimit
	if limit <= 0 {
		limit = 2 << 20
	}
	return io.ReadAll(http.MaxBytesReader(ctx.Writer, ctx.Request.Body, limit))
}

func (req *createAutomationArticleDraftRequest) normalize() {
	req.Title = strings.TrimSpace(req.Title)
	req.Summary = strings.TrimSpace(req.Summary)
	req.Content = strings.TrimSpace(req.Content)
	req.Slug = strings.TrimSpace(req.Slug)
	req.Cover = strings.TrimSpace(req.Cover)
	req.SourceTopic = strings.TrimSpace(req.SourceTopic)
	req.SourcePrompt = strings.TrimSpace(req.SourcePrompt)
	req.GenerationModel = strings.TrimSpace(req.GenerationModel)
}

func (req createAutomationArticleDraftRequest) validate() error {
	switch {
	case req.Title == "":
		return fmt.Errorf("title is required")
	case len([]rune(req.Title)) > 160:
		return fmt.Errorf("title is too long")
	case req.Summary == "":
		return fmt.Errorf("summary is required")
	case len([]rune(req.Summary)) > 500:
		return fmt.Errorf("summary is too long")
	case req.Content == "":
		return fmt.Errorf("content is required")
	case req.Slug != "" && !automationSlugPattern.MatchString(req.Slug):
		return fmt.Errorf("slug is invalid")
	case len([]rune(req.SourceTopic)) > 200:
		return fmt.Errorf("source topic is too long")
	case len([]rune(req.SourcePrompt)) > 5000:
		return fmt.Errorf("source prompt is too long")
	case len([]rune(req.GenerationModel)) > 200:
		return fmt.Errorf("generation model is too long")
	default:
		return nil
	}
}

func (req createAutomationArticleDraftRequest) checkOutdated() bool {
	if req.CheckOutdated == nil {
		return true
	}
	return *req.CheckOutdated
}

func (server *Server) handleAutomationDraftReplay(ctx *gin.Context, existingRequest db.AutomationArticleRequest, requestHash string, idempotencyKey string) {
	if existingRequest.RequestHash != requestHash {
		ctx.JSON(http.StatusConflict, errorResponse(fmt.Errorf("idempotency key reused with different request body")))
		return
	}
	if existingRequest.Status != automationRequestStatusCreated || !existingRequest.ArticleID.Valid {
		ctx.JSON(http.StatusConflict, errorResponse(fmt.Errorf("idempotency request is not replayable")))
		return
	}

	articleID := uuid.UUID(existingRequest.ArticleID.Bytes)
	ctx.JSON(http.StatusOK, server.newAutomationDraftResponse(articleID, existingRequest.Title, idempotencyKey, "replayed"))
}

func (server *Server) recordAutomationDraftFailure(
	ctx *gin.Context,
	req createAutomationArticleDraftRequest,
	status string,
	requestHash string,
	idempotencyKey string,
	keyID string,
	failure error,
) error {
	_, err := server.store.CreateAutomationArticleRequest(ctx, db.CreateAutomationArticleRequestParams{
		IdempotencyKey:  idempotencyKey,
		RequestHash:     requestHash,
		KeyID:           keyID,
		Status:          status,
		Title:           req.Title,
		SourceTopic:     req.SourceTopic,
		SourcePrompt:    req.SourcePrompt,
		GenerationModel: req.GenerationModel,
		ErrorMessage:    failure.Error(),
		ClientIp:        ctx.ClientIP(),
		UserAgent:       ctx.Request.UserAgent(),
	})
	if err != nil {
		return err
	}

	if server.taskDistributor == nil || server.automationNotifyEmail() == "" {
		return nil
	}
	return server.taskDistributor.DistributeTaskNotifyAutomationDraft(ctx, &worker.PayloadNotifyAutomationDraft{
		Kind:            "failure",
		Title:           req.Title,
		IdempotencyKey:  idempotencyKey,
		GenerationModel: req.GenerationModel,
		ErrorMessage:    failure.Error(),
		NotifyEmail:     server.automationNotifyEmail(),
	})
}

func (server *Server) enqueueAutomationDraftSuccess(
	ctx *gin.Context,
	req createAutomationArticleDraftRequest,
	categoryName string,
	articleID uuid.UUID,
	reviewURL string,
	idempotencyKey string,
) error {
	if server.taskDistributor == nil {
		return fmt.Errorf("automation notification distributor is not configured")
	}
	notifyEmail := server.automationNotifyEmail()
	if notifyEmail == "" {
		return fmt.Errorf("automation notify email is not configured")
	}
	return server.taskDistributor.DistributeTaskNotifyAutomationDraft(ctx, &worker.PayloadNotifyAutomationDraft{
		Kind:            "success",
		ArticleID:       articleID,
		Title:           req.Title,
		CategoryName:    categoryName,
		ReviewURL:       reviewURL,
		IdempotencyKey:  idempotencyKey,
		GenerationModel: req.GenerationModel,
		NotifyEmail:     notifyEmail,
	})
}

func (server *Server) newAutomationDraftResponse(articleID uuid.UUID, title string, idempotencyKey string, status string) automationDraftResponse {
	return automationDraftResponse{
		Article: automationDraftArticleResponse{
			ID:                  articleID,
			Title:               title,
			IsPublish:           false,
			CreatedByAutomation: true,
			AutomationStatus:    "pending_review",
		},
		ReviewURL:      server.automationReviewURL(articleID),
		IdempotencyKey: idempotencyKey,
		Status:         status,
	}
}

func (server *Server) automationDraftAPIEnabled() bool {
	return server.config.AutomationHMACKeyID != "" && server.config.AutomationHMACSecret != ""
}

func (server *Server) automationDailyDraftLimit() int64 {
	if server.config.AutomationDailyDraftLimit <= 0 {
		return 1
	}
	return server.config.AutomationDailyDraftLimit
}

func (server *Server) automationNotifyEmail() string {
	if server.config.AutomationNotifyEmail != "" {
		return server.config.AutomationNotifyEmail
	}
	return server.config.EmailSenderAddress
}

func (server *Server) automationReviewURL(articleID uuid.UUID) string {
	return fmt.Sprintf("%s/backend/articles/%s", strings.TrimRight(server.config.Domain, "/"), articleID.String())
}

func calculateAutomationReadTime(htmlContent string) string {
	plainText := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(htmlContent, "")
	wordCount := len([]rune(plainText))
	if wordCount <= 400 {
		return "1 分钟"
	}
	minutes := (wordCount + 399) / 400
	return fmt.Sprintf("%d 分钟", minutes)
}
