package db

import (
	"context"

	"github.com/google/uuid"
)

type DisableVisitorUserTxParams struct {
	ID             uuid.UUID `json:"id"`
	DisabledReason string    `json:"disabled_reason"`
}

type DisableVisitorUserTxResult struct {
	User User `json:"user"`
}

func (store *SQLStore) DisableVisitorUserTx(ctx context.Context, arg DisableVisitorUserTxParams) (DisableVisitorUserTxResult, error) {
	var result DisableVisitorUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		user, err := q.DisableVisitorUser(ctx, DisableVisitorUserParams{
			ID:             arg.ID,
			DisabledReason: arg.DisabledReason,
		})
		if err != nil {
			return err
		}

		if err := q.BlockUserSessions(ctx, arg.ID); err != nil {
			return err
		}

		result.User = user
		return nil
	})

	return result, err
}
