package errs

import (
	"errors"
	"fmt"
)

var (
	// ErrPointerOnly 只支持一级指针作为输入
	// 看到这个 error 说明你输入了其它的东西
	// 我们并不希望用户能够直接使用 err == ErrPointerOnly
	// 所以放在我们的 internal 包里
	ErrPointerOnly               = errors.New("orm: 只支持一级指针作为输入，例如：*User")
	ErrNoRows                    = errors.New("orm: 未找到数据")
	ErrNoUpdatedColumns          = errors.New("orm: 未指定更新的列")
	ErrTooManyReturnedColumns    = errors.New("orm: 过多列")
	ErrInsertZeroRow             = errors.New("orm: 插入 0 行")
	ErrUnknownColumn             = errors.New("orm: 未知列")
	ErrUnknownField              = errors.New("orm: 未知字段")
	ErrUnsupportedAssignableType = errors.New("orm: 不支持的赋值类型")
)

// NewErrUnknownField 创建并返回一个错误，用于指示传入的字段是一个未知字段
func NewErrUnknownField(fd string) error {
	return fmt.Errorf("orm: 未知字段 %s", fd)
}

// NewErrUnknownColumn 创建并返回一个表示未知列错误的error对象
func NewErrUnknownColumn(col string) error {
	return fmt.Errorf("orm: 未知列 %s", col)
}

// NewErrUnsupportedAssignableType 创建一个错误，用于表示不支持的可分配类型
func NewErrUnsupportedAssignableType(exp any) error {
	return fmt.Errorf("orm: 不支持的 Assignable 表达式 %v", exp)
}

// NewErrUnsupportedExpressionType 创建一个错误，用于指示不支持的表达式类型
func NewErrUnsupportedExpressionType(exp any) error {
	return fmt.Errorf("orm: 不支持的表达式 %v", exp)
}

// NewErrUnsupportedTableType 创建一个错误，用于指示不支持的表引用类型
func NewErrUnsupportedTableType(exp any) error {
	return fmt.Errorf("orm: 不支持的 TableReference %v", exp)
}

// NewErrUnsupportedSelectable 创建并返回一个错误，用于表示不支持的可选择项
// 这个函数通常用于处理 ORM (对象关系映射) 操作中不被支持的目标列情况
func NewErrUnsupportedSelectable(exp any) error {
	return fmt.Errorf("orm: 不支持的目标列 %v", exp)
}

// 后面可以考虑支持错误码
// func NewErrUnsupportedExpressionType(exp any) error {
// 	return fmt.Errorf("orm-50001: 不支持的表达式 %v", exp)
// }

// 后面还可以考虑用 AST 分析源码，生成错误排除手册，例如
// @ErrUnsupportedExpressionType 40001
// 发生该错误，主要是因为传入了不支持的 Expression 的实际类型
// 一般来说，这是因为中间件

// NewErrInvalidTagContent 创建并返回一个错误，用于指示提供的标签内容无效
func NewErrInvalidTagContent(tag string) error {
	return fmt.Errorf("orm: 错误的标签设置: %s", tag)
}

// NewErrFailToRollbackTx 创建一个新的错误，用于表示事务回滚失败的情况
// 参数:
// - bizErr: 业务操作中发生的错误。这个错误会被包装在返回的错误中，用于追溯原始错误
// - rbErr: 执行回滚操作时发生的错误。这个错误将被转换为字符串并包含在返回的错误信息中
// - panicked: 一个布尔值，表示在错误发生时是否触发了panic。这用于提供更多的上下文信息，以帮助调试
// 返回值:
// - 返回一个自定义的错误，包含了业务错误、回滚错误以及是否发生panic的信息。这样的设计使得错误处理更加详细和明确
func NewErrFailToRollbackTx(bizErr error, rbErr error, panicked bool) error {
	return fmt.Errorf("orm: 回滚事务失败, 业务错误 %w, 回滚错误 %s, panic: %t", bizErr, rbErr.Error(), panicked)
}
