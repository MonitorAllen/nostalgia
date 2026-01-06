package db

import (
	"context"
)

type DeleteCategoryTxParams struct {
	ID          int64
	AfterDelete func() error
}

func (store *SQLStore) DeleteCategoryTx(ctx context.Context, arg DeleteCategoryTxParams) error {
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		err = q.SetArticleDefaultCategoryIdByCategoryId(ctx, arg.ID)
		if err != nil {
			return err
		}

		err = q.DeleteCategory(ctx, arg.ID)
		if err != nil {
			return err
		}

		return arg.AfterDelete()
	})

	return err
}
