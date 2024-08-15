package sorm

import (
	"context"
	"database/sql"
)

// 确保Tx和DB实现了session接口
var _ session = &Tx{}
var _ session = &DB{}

// session 代表一个抽象的概念，即会话
// 暂时做成私有的，后面考虑重构，因为这个东西用户可能有点难以理解
type session interface {
	// 获取与会话的核心信息
	getCore() core
	// 执行SQL查询的封装并返回结果
	queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	// 数据库交互提供基础的操作接口
	execContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Tx struct {
	tx *sql.Tx
	db *DB
	// 事务扩散方案里面，
	// 这个要在 commit 或者 rollback 的时候修改为 true
	// done bool
}

// getCore 返回Tx对象内部字段*DB结构体里的core核心组件
func (t *Tx) getCore() core {
	return t.db.core
}

// queryContext 对事物内部QueryContext进行封装
func (t *Tx) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

// execContext 对事务内部ExecContext进行封装
func (t *Tx) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

// Commit 提交事务
func (t *Tx) Commit() error {
	return t.tx.Commit()
}

// Rollback 回滚事务
func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

// RollbackIfNotCommit 如果事务没有提交，则回滚事务
func (t *Tx) RollbackIfNotCommit() error {
	err := t.tx.Rollback()
	if err != sql.ErrTxDone {
		return err
	}
	return nil
}
