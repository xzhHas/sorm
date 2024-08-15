package sorm

// RawExpr 代表一个原生的 SQL 表达式
// 原生表达式不会被 ORM 框架处理，直接按原样使用
type RawExpr struct {
	raw  string // 原生SQL表达式字符串
	args []any  // 表达式中的参数
}

// selectedAlias 返回空字符串，因为 RawExpr 不使用别名
func (r RawExpr) selectedAlias() string {
	return ""
}

// fieldName 返回空字符串，因为 RawExpr 不是列
func (r RawExpr) fieldName() string {
	return ""
}

// target 返回 nil，因为 RawExpr 不涉及表引用
func (r RawExpr) target() TableReference {
	return nil
}

func (r RawExpr) assign() {}

func (r RawExpr) expr() {}

// AsPredicate 将 RawExpr 转换为 Predicate 对象
// 用于将原生表达式作为谓词条件
func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}

// Raw 创建一个 RawExpr 对象
// expr 是原生 SQL 表达式，args 是表达式中的参数
func Raw(expr string, args ...any) RawExpr {
	return RawExpr{
		raw:  expr,
		args: args,
	}
}

// binaryExpr 代表一个二元表达式
// 包括两个操作数和一个操作符
type binaryExpr struct {
	left  Expression
	op    op
	right Expression
}

func (binaryExpr) expr() {}

// MathExpr 代表一个数学表达式，它是 binaryExpr 的别名
type MathExpr binaryExpr

// Add 创建一个 MathExpr 对象，表示当前表达式加上一个值
func (m MathExpr) Add(val interface{}) MathExpr {
	return MathExpr{
		left:  m,
		op:    opAdd,
		right: valueOf(val),
	}
}

// Multi 创建一个 MathExpr 对象，表示当前表达式乘以一个值
func (m MathExpr) Multi(val interface{}) MathExpr {
	return MathExpr{
		left:  m,
		op:    opMulti,
		right: valueOf(val),
	}
}

func (m MathExpr) expr() {}

// SubqueryExpr 注意，这个谓词这种不是在所有的数据库里面都支持的
// 这里采取的是和 Upsert 不同的做法
// Upsert 里面我们是属于用 dialect 来区别不同的实现
// 这里我们采用另外一种方案，就是直接生成，依赖于数据库来报错
// 实际中两种方案你可以自由替换
type SubqueryExpr struct {
	// 子查询
	s Subquery
	// 谓词，ALL，ANY 或者 SOME
	pred string
}

func (SubqueryExpr) expr() {}

// Any 创建一个 SubqueryExpr 对象，表示子查询的谓词是 ANY
func Any(sub Subquery) SubqueryExpr {
	return SubqueryExpr{
		s:    sub,
		pred: "ANY",
	}
}

// All 创建一个 SubqueryExpr 对象，表示子查询的谓词是 ALL
func All(sub Subquery) SubqueryExpr {
	return SubqueryExpr{
		s:    sub,
		pred: "ALL",
	}
}

// Some 创建一个 SubqueryExpr 对象，表示子查询的谓词是 SOME
func Some(sub Subquery) SubqueryExpr {
	return SubqueryExpr{
		s:    sub,
		pred: "SOME",
	}
}
