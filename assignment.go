package sorm

// Assignable 标记接口，
// 实现该接口意味着可以用于赋值语句，
// 用于在 UPDATE 和 UPSERT 中
type Assignable interface {
	assign()
}

type Assignment struct {
	column string
	val    Expression
}

// Assign 用于创建一个Assignment对象，它代表数据库表中列的赋值操作
// 该函数接受一个列名和一个值，值可以是任意类型
// 如果值是一个Expression接口的实现，直接使用该值；否则，将其包装成一个value对象
// column：要赋值的列名，val：要赋给列的值，可以是任意类型
func Assign(column string, val any) Assignment {
	v, ok := val.(Expression)
	if !ok {
		v = value{val: val}
	}
	return Assignment{
		column: column,
		val:    v,
	}
}

// assign 方法是 Assignment 类的一个私有方法(所以在这里func(a Assignment)assign()就没有使用指针，因为私有方法不能修改)
// 它负责执行具体的分配逻辑，该逻辑在 Assignment 类的实例化对象中被调用
// 该方法目前没有参数和返回值，它的作用是封装分配相关的操作，增强代码的模块化和可维护性
func (a Assignment) assign() {}
