package db

import (
	"context"

	"github.com/google/uuid"
)

// DeleteArticleTxParams contains the input parameters of the transfer transaction
type DeleteArticleTxParams struct {
	ID          uuid.UUID
	AfterUpdate func(articleID uuid.UUID) error
}

func (store *SQLStore) DeleteArticleTx(ctx context.Context, arg DeleteArticleTxParams) error {
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		err = q.DeleteCommentsByArticleID(ctx, arg.ID)
		if err != nil {
			return err
		}

		err = q.DeleteArticle(ctx, arg.ID)
		if err != nil {
			return err
		}

		return arg.AfterUpdate(arg.ID)
	})

	return err
}
