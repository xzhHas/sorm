package sorm

// op 代表操作符
type op string

// 后面可以每次支持新的操作符就加一个
const (
	opEQ    = "="
	opLT    = "<"
	opGT    = ">"
	opIN    = "IN"
	opExist = "EXIST"
	opAND   = "AND"
	opOR    = "OR"
	opNOT   = "NOT"
	opAdd   = "+"
	opMulti = "*"
)

func (o op) String() string {
	return string(o)
}

// Expression 代表语句，或者语句的部分
// 暂时没想好怎么设计方法，所以直接做成标记接口
type Expression interface {
	expr()
}

// exprOf 将给定值转换为 Expression 类型
func exprOf(e any) Expression {
	switch exp := e.(type) {
	case Expression:
		return exp
	default:
		return valueOf(exp)
	}
}

// Predicate 代表一个查询条件
// Predicate 可以通过和 Predicate 组合构成复杂的查询条件
type Predicate binaryExpr

func (Predicate) expr() {}

// Exist 构建一个存在子查询的谓词
// 存在子查询用于测试一个子查询是否返回至少一行数据
// 参数 sub 是一个 Subquery 类型的对象，代表一个子查询
// 返回值是一个 Predicate 类型的对象，具体来说是一个表示存在子查询的谓词
func Exist(sub Subquery) Predicate {
	return Predicate{
		op:    opExist,
		right: sub,
	}
}

// Not 返回一个Predicate，其值为传入Predicate的逻辑非
// 这个函数构造了一个新的Predicate实例，其操作类型为NOT，右侧操作数为传入的Predicate
func Not(p Predicate) Predicate {
	return Predicate{
		op:    opNOT,
		right: p,
	}
}

// And 方法用于将两个 Predicate 条件进行逻辑与操作，返回一个新的 Predicate 对象
// 这个方法接收一个 Predicate 类型的参数 r，并返回一个新的 Predicate，其中包含了进行逻辑与操作的两个条件
// 一个 Predicate 对象，将与调用实例进行逻辑与操作
// 返回一个新的 Predicate 对象，表示两个条件通过 AND 操作符连接在一起的结果
func (p Predicate) And(r Predicate) Predicate {
	return Predicate{
		left:  p,
		op:    opAND,
		right: r,
	}
}

// Or 方法将当前的 Predicate 实例与另一个 Predicate 实例进行“或”操作
// 这个方法支持 Predicate 实例之间的逻辑“或”操作，用于构建更复杂的查询条件
// 参数 right 是要与当前 Predicate 实例进行“或”操作的另一个 Predicate 实例
// 返回值是一个新的 Predicate 实例，表示左操作数（当前实例）与右操作数（参数 right）的“或”操作结果
func (p Predicate) Or(right Predicate) Predicate {
	return Predicate{
		left:  p,
		op:    opOR,
		right: right,
	}
}
