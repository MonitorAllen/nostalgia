package db

import (
	"context"
)

// CreateCategoryTxParams contains the input parameters of the create category transaction
type CreateCategoryTxParams struct {
	Name        string
	AfterCreate func() error
}

// CreateCategoryTxResult is the result of the create category transaction
type CreateCategoryTxResult struct {
	Category Category
}

func (store *SQLStore) CreateCategoryTx(ctx context.Context, arg CreateCategoryTxParams) (CreateCategoryTxResult, error) {
	var result CreateCategoryTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Category, err = q.CreateCategory(ctx, arg.Name)
		if err != nil {
			return err
		}

		return arg.AfterCreate()
	})

	return result, err
}
