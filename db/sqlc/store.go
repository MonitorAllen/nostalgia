package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides all functions to execute db queries and transactions
type Store interface {
	Querier
	Ping(ctx context.Context) error
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error)
	CreateAutomationArticleTx(ctx context.Context, arg CreateAutomationArticleTxParams) (CreateAutomationArticleTxResult, error)
	VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error)
	UpdateArticleTx(ctx context.Context, arg UpdateArticleTxParams) (UpdateArticleTxResult, error)
	DeleteArticleTx(ctx context.Context, arg DeleteArticleTxParams) error
	DeleteCategoryTx(ctx context.Context, arg DeleteCategoryTxParams) error
	UpdateCategoryTx(ctx context.Context, arg UpdateCategoryTxParams) (UpdateCategoryTxResult, error)
	CreateCategoryTx(ctx context.Context, arg CreateCategoryTxParams) (CreateCategoryTxResult, error)
}

// SQLStore provides all functions to execute SQL queries and transactions
type SQLStore struct {
	connPool *pgxpool.Pool
	*Queries
}

func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}
}

func (store *SQLStore) Ping(ctx context.Context) error {
	return store.connPool.Ping(ctx)
}
