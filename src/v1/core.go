package v1

import (
	"context"
	"database/sql"
	"sorm/src/v1/internal/valuer"
	"sorm/src/v1/model"
)

// core 作为 orm 库的核心组件设计，封装一些基础服务和配置，以支持更高级别的数据库交互操作
type core struct {
	// r 是模型注册表，用于管理和注册 ORM 中使用的模型。
	// 它跟踪所有注册的模型类型及其与数据库表的映射关系。
	// 通过模型注册表，ORM 可以将数据库表中的数据映射到应用程序中的对象，
	// 并将对象的更改同步回数据库。
	r model.Registry
	// dialect 是数据库方言，定义了特定数据库系统的 SQL 语法和行为规则。
	// 不同的数据库系统（如 MySQL、PostgreSQL、SQLite 等）有不同的方言。
	// 通过指定方言，ORM 可以生成与目标数据库兼容的 SQL 语句。
	dialect Dialect
	// valCreator 是值创建器，用于创建和管理 ORM 实体中的字段值。
	// 它处理数据类型转换，将数据库查询结果转换为应用程序中的值类型，
	// 或者将应用程序中的值类型转换为数据库可接受的格式。
	valCreator valuer.Creator
	// ms 是中间件数组，用于在数据库操作过程中处理拦截和增强功能。
	// 中间件可以实现日志记录、事务管理、权限检查等功能。
	// 通过将中间件注册到 core 结构体中，可以在执行数据库操作之前和之后插入自定义逻辑。
	ms []Middleware
}

// getHandler 根据提供的查询上下文执行数据库查询，并将结果映射到指定的结构体类型 T。
// 这个函数展示了如何构建查询、执行查询，并将查询结果映射到一个泛型类型 T 的实例。
// 参数:
// - ctx: 上下文对象，用于控制请求的生命周期。
// - sess: 数据库会话对象，用于执行查询。
// - c: 包含映射和值创建逻辑的核心对象。
// - qc: 包含查询构建器和上下文信息的查询上下文对象。
// 返回值:
// - *QueryResult: 包含查询结果和可能的错误信息的指针。
func getHandler[T any](ctx context.Context, sess session, c core, qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	rows, err := sess.queryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	if !rows.Next() {
		return &QueryResult{
			Err: ErrNoRows,
		}
	}

	tp := new(T)
	meta, err := c.r.Get(tp)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	val := c.valCreator(tp, meta)
	err = val.SetColumns(rows)
	return &QueryResult{
		Result: tp,
		Err:    err,
	}
}

// get 函数用于执行查询操作，它支持泛型参数 T，可以处理不同类型的查询请求。
// 参数:
//   - ctx: 上下文对象，用于传递请求相关的状态和配置。
//   - c: 核心逻辑对象，包含查询所需的必要信息。
//   - sess: 会话对象，用于管理与数据库的交互。
//   - qc: 查询上下文，包含查询的详细配置和状态。
//
// 返回值:
//   - *QueryResult: 查询结果的指针，包含查询操作的结果数据和状态。
func get[T any](ctx context.Context, c core, sess session, qc *QueryContext) *QueryResult {
	var handler HandleFunc = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return getHandler[T](ctx, sess, c, qc)
	}
	// 获取中间件列表，中间件用于在查询处理前后执行额外的逻辑
	ms := c.ms
	// 逆序遍历中间件列表，将每个中间件的处理逻辑与当前处理器组合
	for i := len(ms) - 1; i >= 0; i-- {
		handler = ms[i](handler)
	}
	return handler(ctx, qc)
}

// exec 执行一个数据库查询操作。
// 它接受一个上下文对象，一个数据库会话，一个核心处理对象以及一个查询上下文作为参数。
// 返回值包含查询结果和可能的错误信息。
func exec(ctx context.Context, sess session, c core, qc *QueryContext) Result {
	var handler HandleFunc = func(ctx context.Context, qc *QueryContext) *QueryResult {
		q, err := qc.Builder.Build()
		if err != nil {
			return &QueryResult{
				Err: err,
			}
		}
		res, err := sess.execContext(ctx, q.SQL, q.Args...)
		return &QueryResult{Err: err, Result: res}
	}
	// 获取中间件列表
	ms := c.ms
	// 通过中间件链反转的顺序将处理函数与中间件结合
	for i := len(ms) - 1; i >= 0; i-- {
		handler = ms[i](handler)
	}
	qr := handler(ctx, qc)
	var res sql.Result
	if qr.Result != nil {
		res = qr.Result.(sql.Result)
	}
	return Result{err: qr.Err, res: res}
}
