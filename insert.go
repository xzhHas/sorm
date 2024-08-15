package sorm

import (
	"context"
	"github.com/xzhHas/sorm/internal/errs"
	"github.com/xzhHas/sorm/model"
)

// UpsertBuilder (加载拦截器)定义了一个用于构建 upsert 操作的对象
// 其中 T 是一个泛型类型，代表可以插入的记录类型
// i 指向了一个 Inserter 对象，用于最终执行 upsert 操作
// conflictColumns 存储了用于判断冲突的列名
type UpsertBuilder[T any] struct {
	i               *Inserter[T]
	conflictColumns []string
}

// Upsert 结构体定义了 upsert 操作的具体细节
// conflictColumns 用于存储在 upsert 操作中用于判断冲突的列名
// assigns 存储了在发生冲突时要更新的列及其新值
type Upsert struct {
	conflictColumns []string
	assigns         []Assignable
}

// ConflictColumns 方法用于指定在执行 upsert 操作时，哪些列用于判断冲突
func (o *UpsertBuilder[T]) ConflictColumns(cols ...string) *UpsertBuilder[T] {
	o.conflictColumns = cols
	return o
}

// Update 也可以看做是一个终结方法，重新回到 Inserter 里面
func (o *UpsertBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.upsert = &Upsert{
		conflictColumns: o.conflictColumns,
		assigns:         assigns,
	}
	return o.i
}

// Inserter 定义了一个用于执行插入操作的对象
// 其中 T 是一个泛型类型，代表可以插入的记录类型。
type Inserter[T any] struct {
	builder
	values  []*T     // values 存储了待插入的数据记录
	columns []string // columns 存储了待插入数据的列名
	upsert  *Upsert  // upsert 存储了 upsert 操作的详细信息
	sess    session  // sess 是与数据库交互的会话对象
}

// NewInserter 创建一个新的 Inserter 实例
func NewInserter[T any](sess session) *Inserter[T] {
	c := sess.getCore()
	return &Inserter[T]{
		sess: sess,
		builder: builder{
			core:    c,
			dialect: c.dialect,
			quoter:  c.dialect.quoter(),
		},
	}
}

// Values 设置要插入的数据
func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

// OnDuplicateKey 处理主键冲突的情况（即当发生主键冲突时的行为）
func (i *Inserter[T]) OnDuplicateKey() *UpsertBuilder[T] {
	return &UpsertBuilder[T]{
		i: i,
	}
}

// Fields 指定要插入的列
// TODO 目前我们只支持指定具体的列，但是不支持复杂的表达式
// 例如不支持 VALUES(..., now(), now()) 这种在 MySQL 里面常用的
func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
	i.columns = cols
	return i
}

// Build 构建 SQL 插入语句
func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}
	m, err := i.r.Get(i.values[0])
	i.model = m
	if err != nil {
		return nil, err
	}
	i.sb.WriteString("INSERT INTO ")
	i.quote(m.TableName)
	i.sb.WriteString("(")

	fields := m.Fields
	if len(i.columns) != 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, c := range i.columns {
			field, ok := m.FieldMap[c]
			if !ok {
				return nil, errs.NewErrUnknownField(c)
			}
			fields = append(fields, field)
		}
	}

	// (len(i.values) + 1) 中 +1 是考虑到 UPSERT 语句会传递额外的参数
	i.args = make([]any, 0, len(fields)*(len(i.values)+1))
	for idx, fd := range fields {
		if idx > 0 {
			i.sb.WriteByte(',')
		}
		i.quote(fd.ColName)
	}

	i.sb.WriteString(") VALUES")
	for vIdx, val := range i.values {
		if vIdx > 0 {
			i.sb.WriteByte(',')
		}
		refVal := i.valCreator(val, m)
		i.sb.WriteByte('(')
		for fIdx, field := range fields {
			if fIdx > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')
			fdVal, err := refVal.Field(field.GoName)
			if err != nil {
				return nil, err
			}
			i.addArgs(fdVal)
		}
		i.sb.WriteByte(')')
	}

	if i.upsert != nil {
		err = i.core.dialect.buildUpsert(&i.builder, i.upsert)
		if err != nil {
			return nil, err
		}
	}

	i.sb.WriteString(";")
	return &Query{
		SQL:  i.sb.String(),
		Args: i.args,
	}, nil
}

// Exec 执行插入操作
func (i *Inserter[T]) Exec(ctx context.Context) Result {
	return exec(ctx, i.sess, i.core, &QueryContext{
		Builder: i,
		Type:    "INSERT",
	})
}
