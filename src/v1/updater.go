package v1

import (
	"context"
	"sorm/src/v1/internal/errs"
)

// Updater 结构体表示一个用于执行数据库更新操作的对象
// 参数 T 表示要更新的数据类型
type Updater[T any] struct {
	// 内嵌 builder 结构体提供构建 SQL 查询语句的能力
	builder
	// assigns 存储要更新的字段及其对应的值
	assigns []Assignable
	// val 指向要更新的具体数据类型的指针
	val *T
	// where 存储更新操作的 WHERE 条件
	where []Predicate
	// sess 是与数据库交互的会话对象
	sess session
}

// NewUpdater 创建并返回一个新的 Updater 实例
// - sess: 一个 session 对象，用于获取核心(core)和会话信息(session)
// - *Updater[T]: 返回一个初始化了的 Updater 实例，准备好进行更新操作
func NewUpdater[T any](sess session) *Updater[T] {
	c := sess.getCore()
	return &Updater[T]{
		builder: builder{
			core:    c,
			dialect: c.dialect,
			quoter:  c.dialect.quoter(),
		},
		sess: sess,
	}
}

// Update 方法用于更新 Updater 中的值
// 该方法通过传入的参数 t 更新 Updater 内部的状态，并返回 Updater 实例本身以支持链式调用
// t *T: 要更新的值，可以是任何实现了指针类型的泛型 T
// *Updater[T]: 返回更新后的 Updater 实例指针
func (u *Updater[T]) Update(t *T) *Updater[T] {
	u.val = t
	return u
}

// Set 方法用于设置更新器将要更新的值
// 该方法接收一个或多个可分配的值，并将它们存储在 Updater 结构体中，以便后续更新操作使用
// 通过返回 *Updater[T] 类型的指针，该方法支持方法链调用，允许在设置值之后立即进行其他操作
// assigns: 一个或多个实现了 Assignable 接口的值，表示可以被赋值到内部状态的值
// 返回 Updater 类型的指针，允许进行方法链调用
func (u *Updater[T]) Set(assigns ...Assignable) *Updater[T] {
	u.assigns = assigns
	return u
}

// Build 方法用于构建更新语句
func (u *Updater[T]) Build() (*Query, error) {
	// 检查是否有更新的列
	if len(u.assigns) == 0 {
		return nil, errs.ErrNoUpdatedColumns
	}
	// 如果没有提供值，则初始化为T类型的零值
	if u.val == nil {
		u.val = new(T)
	}
	model, err := u.r.Get(u.val)
	if err != nil {
		return nil, err
	}
	u.model = model
	u.sb.WriteString("UPDATE ")
	u.quote(model.TableName)
	u.sb.WriteString(" SET ")
	// 准备更新的列
	val := u.valCreator(u.val, model)
	// 遍历要更新的列，构建 SET 语句
	for i, a := range u.assigns {
		if i > 0 {
			u.sb.WriteByte(',')
		}
		switch assign := a.(type) {
		case Column:
			if err = u.buildColumn(assign.table, assign.name); err != nil {
				return nil, err
			}
			u.sb.WriteString("=?")
			arg, err := val.Field(assign.name)
			if err != nil {
				return nil, err
			}
			u.addArgs(arg)
		case Assignment:
			if err = u.buildAssignment(assign); err != nil {
				return nil, err
			}
		default:
			return nil, errs.NewErrUnsupportedAssignableType(a)
		}
	}
	// 构建WHERE子句
	if len(u.where) > 0 {
		u.sb.WriteString(" WHERE ")
		if err = u.buildPredicates(u.where); err != nil {
			return nil, err
		}
	}
	u.sb.WriteByte(';')
	return &Query{
		SQL:  u.sb.String(),
		Args: u.args,
	}, nil
}

// buildAssignment 构建并添加一个赋值操作到更新语句中
// assign: 表示一个列和值的映射，其中列名和值分别由Assignment类型的column和val字段表示
// 该函数将使用这一映射来构建更新语句的一部分
func (u *Updater[T]) buildAssignment(assign Assignment) error {
	if err := u.buildColumn(nil, assign.column); err != nil {
		return err
	}
	u.sb.WriteByte('=')
	return u.buildExpression(assign.val)
}

// Where 方法用于设置更新操作的条件
// 它允许调用者指定一个或多个谓词来限定哪些记录应该被更新
// 该方法通过返回带有修改过的内部状态的 Updater 实例，支持方法链调用
// ps ...Predicate - 一个或多个谓词，用于确定哪些记录符合条件，应被更新
// *Updater[T] - 返回带有修改过条件设置的 Updater 实例，允许进行链式调用
func (u *Updater[T]) Where(ps ...Predicate) *Updater[T] {
	u.where = ps
	return u
}

// Exec 执行更新操作
// 该方法首先通过调用 Build 方法构建更新查询，然后使用当前会话执行构建的查询
// 它接受一个 context.Context 参数，用于控制请求的超时和取消
func (u *Updater[T]) Exec(ctx context.Context) Result {
	q, err := u.Build()
	if err != nil {
		return Result{err: err}
	}
	res, err := u.sess.execContext(ctx, q.SQL, q.Args...)
	return Result{err: err, res: res}
}
