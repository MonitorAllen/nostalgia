package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateAutomationArticleDraftParams struct {
	ID            uuid.UUID
	Title         string
	Summary       string
	Content       string
	Owner         uuid.UUID
	CategoryID    int64
	Cover         string
	Slug          pgtype.Text
	CheckOutdated bool
	ReadTime      string
}

type CreateAutomationArticleTxParams struct {
	Request CreateAutomationArticleRequestParams
	Article CreateAutomationArticleDraftParams
}

type CreateAutomationArticleTxResult struct {
	Request AutomationArticleRequest
	Article Article
}

func (store *SQLStore) CreateAutomationArticleTx(ctx context.Context, arg CreateAutomationArticleTxParams) (CreateAutomationArticleTxResult, error) {
	var result CreateAutomationArticleTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Request, err = q.CreateAutomationArticleRequest(ctx, arg.Request)
		if err != nil {
			return err
		}

		result.Article, err = q.CreateAutomationArticle(ctx, CreateAutomationArticleParams{
			ID:                  arg.Article.ID,
			Title:               arg.Article.Title,
			Summary:             arg.Article.Summary,
			Content:             arg.Article.Content,
			Owner:               arg.Article.Owner,
			CategoryID:          arg.Article.CategoryID,
			Cover:               arg.Article.Cover,
			Slug:                arg.Article.Slug,
			CheckOutdated:       arg.Article.CheckOutdated,
			ReadTime:            arg.Article.ReadTime,
			AutomationRequestID: pgtype.Int8{Int64: result.Request.ID, Valid: true},
		})
		if err != nil {
			return err
		}

		result.Request, err = q.MarkAutomationArticleRequestCreated(ctx, MarkAutomationArticleRequestCreatedParams{
			ID:        result.Request.ID,
			ArticleID: pgtype.UUID{Bytes: result.Article.ID, Valid: true},
		})
		return err
	})

	return result, err
}
