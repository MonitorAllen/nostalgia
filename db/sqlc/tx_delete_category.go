package db

import (
	"context"
)

func (store *SQLStore) DeleteCategoryTx(ctx context.Context, ID int64) error {
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		err = q.SetArticleDefaultCategoryIdByCategoryId(ctx, ID)
		if err != nil {
			return err
		}

		err = q.DeleteCategory(ctx, ID)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
