package v1

import (
	"context"
	"database/sql"
	"github.com/xzhHas/sorm/src/v1/internal/errs"
)

// Selector 是一个泛型结构体，用于构建和执行数据库查询
// T 是泛型参数，允许 Selector 操作不同类型的实体
type Selector[T any] struct {
	builder
	table   TableReference
	where   []Predicate
	having  []Predicate
	columns []Selectable
	groupBy []Column
	offset  int
	limit   int
	sess    session
}

// Select 方法用于指定查询操作选择的列
// 该方法通过使 Selector 结构体能够链式调用，增强了数据库查询的灵活性和便捷性
// cols 一个可变参数，表示可以被选择的列的接口类型切片
// *Selector[T] 返回 Selector 实例的引用，使得可以进行链式调用
// 这种设计模式使得用户可以一次性设置查询的多个参数，提高代码的可读性和效率
func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
	return s
}

// From 方法用于设置查询语句的起始表，通过链式调用返回 Selector 实例
// 指定表名，如果是空字符串，那么将会使用默认表名
// 参数 tbl 是一个 TableReference 类型，表示查询的起始表
// 返回值是 *Selector[T] 类型，允许调用者继续链式调用其他方法
func (s *Selector[T]) From(tbl TableReference) *Selector[T] {
	s.table = tbl
	return s
}

// Build 构建一个 Query 对象，该对象包含根据 Selector 实例配置的 SQL 查询语句和参数
// T 是一个类型参数，代表查询结果的模型类型
// 返回值是构建好的 Query 对象和可能的错误
func (s *Selector[T]) Build() (*Query, error) {
	var err error
	// 初始化模型，通过反射机制获取T类型的实例
	s.model, err = s.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	// 开始构建 SELECT 语句
	s.sb.WriteString("SELECT ")
	if err = s.buildColumns(); err != nil {
		return nil, err
	}
	// 指定查询的数据表
	s.sb.WriteString(" FROM ")
	if err = s.buildTable(s.table); err != nil {
		return nil, err
	}
	// 构造 WHERE字句，用于过滤条件
	if len(s.where) > 0 {
		// 类似这种可有可无的部分，都要在前面加一个空格
		s.sb.WriteString(" WHERE ")
		if err = s.buildPredicates(s.where); err != nil {
			return nil, err
		}
	}
	// 构造 GROUP BY，用于分组
	if len(s.groupBy) > 0 {
		s.sb.WriteString(" GROUP BY ")
		for i, c := range s.groupBy {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			if err = s.buildColumn(c, false); err != nil {
				return nil, err
			}
		}
	}
	// 构造 HAVING，用于对分组后的数据进行过滤
	if len(s.having) > 0 {
		s.sb.WriteString(" HAVING ")
		if err = s.buildPredicates(s.having); err != nil {
			return nil, err
		}
	}
	// 添加 LIMIT，限制返回结果的数量
	if s.limit > 0 {
		s.sb.WriteString(" LIMIT ?")
		s.addArgs(s.limit)
	}
	// 添加 OFFSET，设置查询结果的其实位置
	if s.offset > 0 {
		s.sb.WriteString(" OFFSET ?")
		s.addArgs(s.offset)
	}

	s.sb.WriteString(";")
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

// buildTable 根据给定的表引用构建查询表部分
// 它处理了不同类型的表引用，并相应地构建查询语句
func (s *Selector[T]) buildTable(table TableReference) error {
	switch tab := table.(type) {
	case nil:
		s.quote(s.model.TableName)
	case Table:
		model, err := s.r.Get(tab.entity)
		if err != nil {
			return err
		}
		s.quote(model.TableName)
		if tab.alias != "" {
			s.sb.WriteString(" AS ")
			s.quote(tab.alias)
		}
	case Join:
		return s.buildJoin(tab)
	case Subquery:
		return s.buildSubquery(tab, true)
	default:
		return errs.NewErrUnsupportedExpressionType(tab)
	}
	return nil
}

// buildJoin 构建一个 JOIN 语句
// tab: Join 类型的参数，包含构建 JOIN 语句所需的所有信息
func (s *Selector[T]) buildJoin(tab Join) error {
	s.sb.WriteByte('(')
	// 构建JOIN的左侧表
	if err := s.buildTable(tab.left); err != nil {
		return err
	}
	// 添加空格和JOIN类型
	s.sb.WriteString(" ")
	s.sb.WriteString(tab.typ)
	s.sb.WriteString(" ")
	// 构建JOIN的右侧表
	if err := s.buildTable(tab.right); err != nil {
		return err
	}
	// 处理 USING 子句
	if len(tab.using) > 0 {
		s.sb.WriteString(" USING (")
		for i, col := range tab.using {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			// 构建USING子句中的列名
			err := s.buildColumn(Column{name: col}, false)
			if err != nil {
				return err
			}
		}
		s.sb.WriteString(")")
	}
	// 如果使用ON关键字，则构建ON子句
	if len(tab.on) > 0 {
		s.sb.WriteString(" ON ")
		err := s.buildPredicates(tab.on)
		if err != nil {
			return err
		}
	}
	s.sb.WriteByte(')')
	return nil
}

// buildColumns构建查询语句中的列部分
// 该方法根据Selector结构体中的columns切片来生成查询语句的列部分
// 切片为空时，会使用通配符*来表示选择所有列。否则，会遍历columns切片中的每个元素
// 根据元素类型分别构建列名、聚合函数、原始表达式或其他支持的可选择项

func (s *Selector[T]) buildColumns() error {
	// 如果columns为空，则使用通配符*，表示选择所有列，并返回nil
	if len(s.columns) == 0 {
		s.sb.WriteByte('*')
		return nil
	}
	for i, c := range s.columns {
		if i > 0 {
			s.sb.WriteByte(',')
		}
		switch val := c.(type) {
		case Column:
			if err := s.buildColumn(val, true); err != nil {
				return err
			}
		case Aggregate:
			if err := s.buildAggregate(val, true); err != nil {
				return err
			}
		case RawExpr:
			s.raw(val)
		default:
			return errs.NewErrUnsupportedSelectable(c)
		}
	}
	return nil
}

// buildColumn 构建查询语句中的列部分
// 该方法负责将指定的列添加到查询语句中，并根据useAlias参数决定是否使用别名
// 它首先调用内部的builder来构建列，然后根据useAlias标志来决定是否构建别名

func (s *Selector[T]) buildColumn(c Column, useAlias bool) error {
	err := s.builder.buildColumn(c.table, c.name)
	if err != nil {
		return err
	}
	if useAlias {
		s.buildAs(c.alias)
	}
	return nil
}

// Where 用于构造 WHERE 查询条件。如果 ps 长度为 0，那么不会构造 WHERE 部分
func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	s.where = ps
	return s
}

// GroupBy 设置 group by 子句
func (s *Selector[T]) GroupBy(cols ...Column) *Selector[T] {
	s.groupBy = cols
	return s
}

func (s *Selector[T]) Having(ps ...Predicate) *Selector[T] {
	s.having = ps
	return s
}

func (s *Selector[T]) Offset(offset int) *Selector[T] {
	s.offset = offset
	return s
}

func (s *Selector[T]) Limit(limit int) *Selector[T] {
	s.limit = limit
	return s
}

// AsSubquery 将当前选择器实例转换为一个子查询
// 此方法允许在更大的查询中使用当前的选择器实例作为子查询
// 它通过创建一个包含选择器实例和指定别名的新子查询对象来实现这一点
func (s *Selector[T]) AsSubquery(alias string) Subquery {
	// 获取当前选择器关联的表，如果没有，则创建一个新的表。
	tbl := s.table
	if tbl == nil {
		tbl = TableOf(new(T))
	}
	// 创建并返回一个新的子查询对象，其中包括当前选择器实例、指定的别名、关联的表和列。
	return Subquery{
		s:       s,
		alias:   alias,
		table:   tbl,
		columns: s.columns,
	}
}

// Get 方法用于从数据库中获取特定类型 T 的数据
// 该方法通过提供的 context.Context 对象来控制请求的取消或超时
func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	// 调用 get 函数执行查询操作，传入上下文对象、core、sess 和查询上下文
	// 这里使用了一个泛型 T，使得同一个函数可以处理不同类型的查询
	res := get[T](ctx, s.core, s.sess, &QueryContext{
		Builder: s,
		Type:    "SELECT",
	})
	// 检查查询结果，如果结果不为空，则返回结果对象和可能的错误
	if res.Result != nil {
		return res.Result.(*T), res.Err
	}
	// 如果结果为空，可能是因为没有找到数据或发生了错误，返回 nil 和可能的错误
	return nil, res.Err
}

// GetMulti 根据提供的上下文从数据库中获取多个实体
// 该方法首先构建查询，然后使用该查询从数据库中检索数据
func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	// 定义一个未指定类型的 SQL.DB 实例，用于执行查询
	var db sql.DB
	// 调用 Build 方法构造一个查询
	q, err := s.Build()
	if err != nil {
		return nil, err
	}
	// 使用上下文执行构造的查询，并获取结果集
	rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		// TODO: 在这里构造 []*T
	}
	// 提醒开发者此处需要实现具体的逻辑来构造返回值
	panic("implement me")
}

// NewSelector 创建并返回一个新的 Selector 实例
// - 该函数使用泛型 T 来允许创建任意类型的 Selector 实例
// - 通过 sess 参数获取数据库操作的核心配置（如连接信息和方言设置）
// - 返回的 Selector 实例封装了 session 和 builder，提供便捷的数据库查询构建方法
func NewSelector[T any](sess session) *Selector[T] {
	c := sess.getCore()
	return &Selector[T]{
		sess: sess,
		builder: builder{
			core:    c,
			dialect: c.dialect,
			quoter:  c.dialect.quoter(),
		},
	}
}

// Selectable 定义了从数据库表中选择数据时需要遵循的方法签名
// - selectedAlias(): 返回选择字段时使用的别名
// - fieldName(): 返回字段名称
// - target(): 返回与此 Selectable 关联的表引用（TableReference），用于确定数据来源
type Selectable interface {
	selectedAlias() string
	fieldName() string
	target() TableReference
}
