package sorm

// Column 代表数据库表中的一个列，包含列的名称、别名和所属表的引用
type Column struct {
	table TableReference //所属表的引用
	name  string         //列的名称
	alias string         //列的别名
}

func (c Column) assign() {}

func (c Column) expr() {}

func (c Column) selectedAlias() string {
	return c.alias
}

func (c Column) fieldName() string {
	return c.name
}

// target 返回 Column 所属的表引用
func (c Column) target() TableReference {
	return c.table
}

// As 创建一个新的 Column 对象，设置其别名
func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
	}
}

// value 代表一个值，用于在 SQL 查询中作为表达式
type value struct {
	val any
}

func (c value) expr() {}

// valueOf 创建一个新的 value 对象
func valueOf(val any) value {
	return value{
		val: val,
	}
}

// C 创建一个新的 Column 对象，设置其名称
func C(name string) Column {
	return Column{name: name}
}

// Add 创建一个 MathExpr 对象，表示当前列加上一个整数增量
func (c Column) Add(delta int) MathExpr {
	return MathExpr{
		left:  c,
		op:    opAdd,
		right: value{val: delta},
	}
}

// Multi 创建一个 MathExpr 对象，表示当前列乘以一个整数增量
func (c Column) Multi(delta int) MathExpr {
	return MathExpr{
		left:  c,
		op:    opAdd,
		right: value{val: delta},
	}
}

// EQ 创建一个 Predicate 对象，表示当前列等于某个值
// EQ 例如 C("id").Eq(12)
func (c Column) EQ(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEQ,
		right: exprOf(arg),
	}
}

// LT 创建一个 Predicate 对象，表示当前列小于某个值
func (c Column) LT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLT,
		right: exprOf(arg),
	}
}

// GT 创建一个 Predicate 对象，表示当前列大于某个值
func (c Column) GT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opGT,
		right: exprOf(arg),
	}
}

// In 创建一个 Predicate 对象，表示当前列的值在给定的多个值中
// In 有两种输入，一种是 IN 子查询
// 另外一种就是普通的值
// 这里我们可以定义两个方法，如 In  和 InQuery，也可以定义一个方法
// 这里我们使用一个方法
func (c Column) In(vals ...any) Predicate {
	return Predicate{
		left:  c,
		op:    opIN,
		right: valueOf(vals),
	}
}

// InQuery 创建一个 Predicate 对象，表示当前列的值在一个子查询的结果中
func (c Column) InQuery(sub Subquery) Predicate {
	return Predicate{
		left:  c,
		op:    opIN,
		right: sub,
	}
}
