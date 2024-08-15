package v1

import (
	"context"
	"github.com/xzhHas/sorm/src/v1/model"
)

// QueryContext 代表执行数据库查询时的上下文信息
type QueryContext struct {
	// Type 声明查询类型。即 SELECT, UPDATE, DELETE 和 INSERT
	Type string
	// builder 使用的时候，大多数情况下你需要转换到具体的类型，才能修改查询
	Builder QueryBuilder
	// Model 存储与查询相关的模型信息
	Model *model.Model
}

// QueryResult 代表数据库查询操作的结果
type QueryResult struct {
	// Result 在不同的查询里面，类型是不同的
	// Selector.Get 里面，这会是单个结果
	// Selector.GetMulti，这会是一个切片
	// 其它情况下，它会是 Result 类型
	Result any
	Err    error
}

// Middleware 定义了一个中间件函数，它接收下一个处理函数作为参数，并返回一个新的处理函数。
// 这种设计模式通常用于在处理请求前后添加额外的功能，例如日志记录、认证等。
type Middleware func(next HandleFunc) HandleFunc

// HandleFunc 定义了处理函数的类型，它接收一个上下文和一个 QueryContext 对象，并返回一个 QueryResult 对象。
// 这个函数用于执行实际的数据库操作。
type HandleFunc func(ctx context.Context, qc *QueryContext) *QueryResult
