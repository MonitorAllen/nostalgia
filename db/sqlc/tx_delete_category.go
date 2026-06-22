package db

import (
	"context"
)

type DeleteCategoryTxParams struct {
	ID                  int64
	DeleteArticles      bool
	AfterDelete         func() error
	AfterDeleteArticles func([]ListArticleResourceRefsByCategoryIDRow) error
}

func (store *SQLStore) DeleteCategoryTx(ctx context.Context, arg DeleteCategoryTxParams) error {
	var articleRefs []ListArticleResourceRefsByCategoryIDRow

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		if arg.DeleteArticles {
			articleRefs, err = q.ListArticleResourceRefsByCategoryID(ctx, arg.ID)
			if err != nil {
				return err
			}

			err = q.DeleteCommentsByCategoryID(ctx, arg.ID)
			if err != nil {
				return err
			}

			err = q.DeleteArticlesByCategoryID(ctx, arg.ID)
			if err != nil {
				return err
			}
		} else {
			err = q.SetArticleDefaultCategoryIdByCategoryId(ctx, arg.ID)
			if err != nil {
				return err
			}
		}

		err = q.DeleteCategory(ctx, arg.ID)
		if err != nil {
			return err
		}

		if arg.AfterDelete != nil {
			return arg.AfterDelete()
		}

		return nil
	})
	if err != nil {
		return err
	}

	if arg.DeleteArticles && arg.AfterDeleteArticles != nil {
		return arg.AfterDeleteArticles(articleRefs)
	}

	return nil
}
