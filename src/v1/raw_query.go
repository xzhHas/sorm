package v1

import (
	"context"
)

var _ Querier[any] = &RawQuerier[any]{}

// RawQuerier 原生查询器
type RawQuerier[T any] struct {
	core
	sess session
	sql  string
	args []any
}

// Exec 执行原生 SQL 语句
func (r *RawQuerier[T]) Exec(ctx context.Context) Result {
	return exec(ctx, r.sess, r.core, &QueryContext{
		Builder: r,
		Type:    "RAW",
	})
}

// Get 获取单条记录
func (r *RawQuerier[T]) Get(ctx context.Context) (*T, error) {
	res := get[T](ctx, r.core, r.sess, &QueryContext{
		Builder: r,
		Type:    "RAW",
	})
	if res.Result != nil {
		return res.Result.(*T), res.Err
	}
	return nil, res.Err
}

// GetMulti 获取多条记录
func (r *RawQuerier[T]) GetMulti(ctx context.Context) ([]*T, error) {
	// TODO implement me
	panic("implement me")
}

// Build 构建 SQL 查询语句
func (r *RawQuerier[T]) Build() (*Query, error) {
	return &Query{
		SQL:  r.sql,
		Args: r.args,
	}, nil
}

// RawQuery 创建一个 RawQuerier 实例
// 泛型参数 T 是目标类型
// 例如，如果查询 User 的数据，那么 T 就是 User
func RawQuery[T any](sess session, sql string, args ...any) *RawQuerier[T] {
	return &RawQuerier[T]{
		sql:  sql,
		args: args,
		core: sess.getCore(),
		sess: sess,
	}
}
