package v1

import (
	"context"
)

// Querier 用于SELLECT 操作
type Querier[T any] interface {
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) ([]*T, error)
}

// Executor 用于INSERT/UPDATE/DELETE 操作
type Executor interface {
	Exec(ctx context.Context) Result
}

type Query struct {
	SQL  string
	Args []any
}

type QueryBuilder interface {
	Build() (*Query, error)
}
