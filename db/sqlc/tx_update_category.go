package db

import (
	"context"
)

type UpdateCategoryTxParams struct {
	UpdateCategoryParams
	AfterUpdate func() error
}

type UpdateCategoryTxResult struct {
	Category Category
}

func (store *SQLStore) UpdateCategoryTx(ctx context.Context, arg UpdateCategoryTxParams) (UpdateCategoryTxResult, error) {
	var result UpdateCategoryTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Category, err = q.UpdateCategory(ctx, arg.UpdateCategoryParams)
		if err != nil {
			return err
		}

		return arg.AfterUpdate()
	})

	return result, err
}
