package sorm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"github.com/xzhHas/sorm/internal/errs"
	"github.com/xzhHas/sorm/internal/valuer"
	"github.com/xzhHas/sorm/model"
	"log"
	"time"
)

type DBOption func(*DB)

type DB struct {
	core
	db *sql.DB
}

// Wait 会等待数据库连接
// 注意只能用于测试
func (db *DB) Wait() error {
	err := db.db.Ping()
	for err == driver.ErrBadConn {
		log.Printf("等待数据库启动...")
		err = db.db.Ping()
		time.Sleep(time.Second)
	}
	return err
}

// Open 创建一个 DB 实例
// driver 也就是驱动的名字，例如：mysql、sqlite3
// dsn 就是数据库的连接字符串
// opts 自定义驱动，例如：分库分表
func Open(driver string, dsn string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	return OpenDB(db, opts...)
}

// OpenDB 用于初始化并返回一个配置好的*DB实例
func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	res := &DB{
		core: core{
			dialect:    MySQL,
			r:          model.NewRegistry(),
			valCreator: valuer.NewUnsafeValue,
		},
		db: db,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

// DBWithDialect 用于指定数据库方言(Open默认设置为MySQL)
func DBWithDialect(dialect Dialect) DBOption {
	return func(db *DB) {
		db.dialect = dialect
	}
}

// DBWithRegistry 用于指定实体与表的映射关系
// 接收一个实现Registry接口的对象，并将其作为数据库的注册表
func DBWithRegistry(r model.Registry) DBOption {
	return func(db *DB) {
		db.r = r
	}
}

// DBUseReflectValuer 使用反射作为值创建器
func DBUseReflectValuer() DBOption {
	return func(db *DB) {
		db.valCreator = valuer.NewReflectValue
	}
}

// DBWithMiddleware 用于指定中间件
func DBWithMiddleware(ms ...Middleware) DBOption {
	return func(db *DB) {
		db.ms = ms
	}
}

// MustNewDB 创建一个 DB，如果失败则会 panic
// 我个人不太喜欢这种
func MustNewDB(driver string, dsn string, opts ...DBOption) *DB {
	db, err := Open(driver, dsn, opts...)
	if err != nil {
		panic(err)
	}
	return db
}

// BeginTx 开启事务
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, db: db}, nil
}

type FN func(ctx context.Context, tx *Tx) error

// DoTx 将会开启事务执行 fn。如果 fn 返回错误或者发生 panic，事务将会回滚，
// 否则提交事务
func (db *DB) DoTx(ctx context.Context, fn FN, opts *sql.TxOptions) (err error) {
	var tx *Tx
	tx, err = db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	panicked := true
	defer func() {
		if panicked || err != nil {
			e := tx.Rollback()
			if e != nil {
				err = errs.NewErrFailToRollbackTx(err, e, panicked)
			}
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(ctx, tx)
	panicked = false
	return err
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	return db.db.Close()
}

// getCore 获取 core
func (db *DB) getCore() core {
	return db.core
}

// queryContext 查询多行数据
func (db *DB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

// execContext 增改删
// 它接收一个context.Context对象用于控制SQL执行的超时和取消
func (db *DB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}
