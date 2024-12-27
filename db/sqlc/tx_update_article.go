package db

import (
	"context"
)

// UpdateArticleTxParams contains the input parameters of the transfer transaction
type UpdateArticleTxParams struct {
	UpdateArticleParams
	AfterUpdate func(article Article) error
}

// UpdateArticleTxResult is the result of the transfer transaction
type UpdateArticleTxResult struct {
	Article Article
}

func (store *SQLStore) UpdateArticleTx(ctx context.Context, arg UpdateArticleTxParams) (UpdateArticleTxResult, error) {
	var result UpdateArticleTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Article, err = q.UpdateArticle(ctx, arg.UpdateArticleParams)
		if err != nil {
			return err
		}

		return arg.AfterUpdate(result.Article)
	})

	return result, err
}
