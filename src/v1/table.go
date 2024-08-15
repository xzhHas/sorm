package v1

// TableReference 是一个接口类型，定义了获取表别名的方法。
// 这个接口被用来确保任何可以参与 JOIN 操作的对象都能够提供一个别名。
type TableReference interface {
	tableAlias() string
}

// Table 结构体表示数据库中的一个表，包含实体和别名信息。
// 这个结构体用于封装表的信息，并且可以用于构建 SQL 查询语句
type Table struct {
	entity any    //表示表中的实体
	alias  string //表的别名
}

// TableOf 创建并返回一个新的 Table 实例。
// 用于从一个实体创建一个 Table 实例，这个实例可以用于构建 SQL 查询语句。
func TableOf(entity any) Table {
	return Table{
		entity: entity,
	}
}

// C 方法返回一个 Column 实例，该实例与当前表关联。
// 用于从当前表中获取一个列的引用，以便于后续的操作如 SELECT 或 JOIN
func (t Table) C(name string) Column {
	return Column{
		name:  name,
		table: t,
	}
}

// tableAlias (表别名)方法实现了 TableReference 接口，返回表的别名。
// 用于获取当前表的别名，这对于构建 SQL 查询时需要的别名非常重要。
func (t Table) tableAlias() string {
	return t.alias
}

// As 方法创建并返回一个新的 Table 实例，并设置其别名。
// 用于给当前表设置别名，这在 SQL 查询中非常常见，尤其是在 JOIN 操作中
func (t Table) As(alias string) Table {
	return Table{
		entity: t.entity,
		alias:  alias,
	}
}

// Join 方法创建并返回一个新的 JoinBuilder 实例，用于构建 JOIN 操作。
// 用于开始构建一个 JOIN 操作，可以指定 JOIN 的类型（默认为 INNER JOIN）
func (t Table) Join(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: target,
		typ:   "JOIN",
	}
}

// LeftJoin 方法创建并返回一个新的 JoinBuilder 实例，用于构建 LEFT JOIN 操作
func (t Table) LeftJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: target,
		typ:   "LEFT JOIN",
	}
}

// RightJoin 方法创建并返回一个新的 JoinBuilder 实例，用于构建 RIGHT JOIN 操作
func (t Table) RightJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: target,
		typ:   "RIGHT JOIN",
	}
}

// JoinBuilder 结构体用于构建 JOIN 操作的中间状态
type JoinBuilder struct {
	left  TableReference //左侧表
	right TableReference //右侧表
	typ   string         //JOIN类型
}

// 这种声明用来确保 JoinBuilder 实现了 TableReference 接口
var _ TableReference = Join{}

// Join 结构体表示一个 JOIN 操作
type Join struct {
	left  TableReference //左侧表
	right TableReference //右侧表
	typ   string         //JOIN类型
	on    []Predicate    //ON子句
	using []string       //USING
}

// Join 方法创建并返回一个新的 JoinBuilder 实例，用于构建 JOIN 操作
func (j Join) Join(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: target,
		typ:   "JOIN",
	}
}

// LeftJoin 方法创建并返回一个新的 JoinBuilder 实例，用于构建 LEFT JOIN 操作
func (j Join) LeftJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: target,
		typ:   "LEFT JOIN",
	}
}

// RightJoin 方法创建并返回一个新的 JoinBuilder 实例，用于构建 RIGHT JOIN 操作
func (j Join) RightJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: target,
		typ:   "RIGHT JOIN",
	}
}

// tableAlias (表别名)方法实现了 TableReference 接口，返回空字符串。
func (j Join) tableAlias() string {
	return ""
}

// On 方法创建并返回一个新的 Join 实例，并设置 ON 子句
func (j *JoinBuilder) On(ps ...Predicate) Join {
	return Join{
		left:  j.left,
		right: j.right,
		on:    ps,
		typ:   j.typ,
	}
}

// Using 方法创建并返回一个新的 Join 实例，并设置 USING 子句
func (j *JoinBuilder) Using(cs ...string) Join {
	return Join{
		left:  j.left,
		right: j.right,
		using: cs,
		typ:   j.typ,
	}
}

// Subquery 结构体表示一个子查询，包含子查询的查询构建器、选择的列和别名。
type Subquery struct {
	s       QueryBuilder   // 使用 QueryBuilder 仅仅是为了让 Subquery 可以是非泛型的
	columns []Selectable   //选择的列
	alias   string         //表别名
	table   TableReference //表
}

// expr 方法实现了 Selectable 接口，返回空值。
func (s Subquery) expr() {}

// selectedAlias 方法实现了 Selectable 接口，返回空值。
func (s Subquery) tableAlias() string {
	return s.alias
}

// Join 方法创建并返回一个新的 JoinBuilder 实例，用于构建 JOIN 操作
func (s Subquery) Join(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: target,
		typ:   "JOIN",
	}
}

// LeftJoin 方法创建并返回一个新的 JoinBuilder 实例，用于构建 LEFT JOIN 操作
func (s Subquery) LeftJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: target,
		typ:   "LEFT JOIN",
	}
}

// RightJoin 方法创建并返回一个新的 JoinBuilder 实例，用于构建 RIGHT JOIN 操作
func (s Subquery) RightJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: target,
		typ:   "RIGHT JOIN",
	}
}

// C 方法创建并返回一个新的 Column 实例，并设置列的名称和表别名。
func (s Subquery) C(name string) Column {
	return Column{
		table: s,
		name:  name,
	}
}
