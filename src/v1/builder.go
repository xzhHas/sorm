package v1

import (
	"sorm/src/v1/internal/errs"
	"sorm/src/v1/model"
	"strings"
)

// builder 用于构建 SQL 查询语句
type builder struct {
	core                    // orm核心组件
	sb      strings.Builder // 高效构建SQL语句时累积和拼接SQL片段
	args    []any           // 存储SQL查询参数的切片
	dialect Dialect         // 数据库方言
	quoter  byte            // 表示SQL标识符(表名和列名)的引号
	model   *model.Model    // model包含结构体和字段信息，用于生成与特定数据表相关的SQL查询
}

// buildColumn 构造列
// 如果 table 没有指定，我们就用 model 来判断列是否存在
func (b *builder) buildColumn(table TableReference, fd string) error {
	var alias string
	if table != nil {
		alias = table.tableAlias()
	}
	if alias != "" {
		b.quote(alias)
		b.sb.WriteByte('.')
	}
	colName, err := b.colName(table, fd)
	if err != nil {
		return err
	}
	b.quote(colName)
	return nil
}

// colName 根据给定的表引用和字段名，返回对应的列名
// 如果无法解析列名，则返回错误
// table - 表引用，可以是 nil、Table 或 Join 类型
// fd - 字段名
func (b *builder) colName(table TableReference, fd string) (string, error) {
	switch tab := table.(type) {
	case nil:
		// 尝试直接从模型的字段映射中获取列
		fdMeta, ok := b.model.FieldMap[fd]
		if !ok {
			return "", errs.NewErrUnknownField(fd)
		}
		return fdMeta.ColName, nil
	case Table:
		// 对于 Table 类型，获取实体的元数据
		m, err := b.r.Get(tab.entity)
		if err != nil {
			return "", err
		}
		fdMeta, ok := m.FieldMap[fd]
		if !ok {
			return "", errs.NewErrUnknownField(fd)
		}
		return fdMeta.ColName, nil
	case Join:
		// 对于 Join 类型，递归地从左右表中查找列
		colName, err := b.colName(tab.left, fd)
		if err != nil {
			return colName, nil
		}
		return b.colName(tab.right, fd)
	case Subquery:
		// 对于 Subquery (子查询)类型，首先检查是否有显式列
		if len(tab.columns) > 0 {
			for _, c := range tab.columns {
				if c.selectedAlias() == fd {
					return fd, nil
				}
				if c.fieldName() == fd {
					return b.colName(c.target(), fd)
				}
			}
			return "", errs.NewErrUnknownField(fd)
		}
		return b.colName(tab.table, fd)
	default:
		return "", errs.NewErrUnsupportedExpressionType(tab)
	}
}

// quote 方法用于将给定的名称用引号包围并添加到构建器中
// 此方法主要用于处理需要被引号包围的标识符，如列名或变量名，以确保在生成的语句中它们被正确地识别和处理
// 该方法直接操作 builder 内部的字符串构建器（StringBuilder）
// 通过在名称前后添加 quoter 字符（引号），来实现包围名称的功能
// 这种方式提高了效率并减少了错误的可能性
func (b *builder) quote(name string) {
	b.sb.WriteByte(b.quoter)
	b.sb.WriteString(name)
	b.sb.WriteByte(b.quoter)
}

// raw 方法用于将给定的 RawExpr 添加到构建器中
func (b *builder) raw(r RawExpr) {
	b.sb.WriteString(r.raw)
	if len(r.args) != 0 {
		b.addArgs(r.args...)
	}
}

// addArgs 向构建器的参数列表中添加新的参数
// 这个方法管理参数切片的初始化和扩展，以有效地支持构建查询语句时的参数附加
func (b *builder) addArgs(args ...any) {
	if b.args == nil {
		// 很少有查询能够超过八个参数
		// INSERT 除外
		b.args = make([]any, 0, 8)
	}
	b.args = append(b.args, args...)
}

// buildPredicates 递归构建一组谓词表达式。
// 它接收一个 Predicate 切片作为参数，并尝试将它们组合成一个单独的谓词，
// 然后将这个组合后的谓词作为表达式进行构建。
// 如果构建过程中发生错误，会返回该错误。
func (b *builder) buildPredicates(ps []Predicate) error {
	p := ps[0]
	for i := 1; i < len(ps); i++ {
		p = p.And(ps[i])
	}
	return b.buildExpression(p)
}

// buildExpression 根据给定的表达式对象构建相应的SQL表达式。
// 如果表达式为nil，则函数返回nil。
// 如果表达式对象是不支持的类型，函数将返回一个错误。
func (b *builder) buildExpression(e Expression) error {
	if e == nil {
		return nil
	}
	switch exp := e.(type) {
	case Column:
		// 当表达式为列时，构建列的SQL表示
		return b.buildColumn(exp.table, exp.name)
	case Aggregate:
		// 当表达式为聚合函数时，构建聚合函数的SQL表示
		return b.buildAggregate(exp, false)
	case value:
		// 当表达式为值时，添加一个问号占位符并记录值到参数列表
		b.sb.WriteByte('?')
		b.addArgs(exp.val)
	case RawExpr:
		// 当表达式为原始SQL时，直接将其添加到构建的SQL中
		b.raw(exp)
	case MathExpr:
		// 当表达式为数学表达式时，构建相应的二元表达式
		return b.buildBinaryExpr(binaryExpr(exp))
	case Predicate:
		// 当表达式为谓词时，构建相应的二元表达式
		return b.buildBinaryExpr(binaryExpr(exp))
	case SubqueryExpr:
		// 当表达式为子查询表达式时，添加谓词和构建子查询
		b.sb.WriteString(exp.pred)
		b.sb.WriteByte(' ')
		return b.buildSubquery(exp.s, false)
	case Subquery:
		// 当表达式为子查询时，构建子查询
		return b.buildSubquery(exp, false)
	case binaryExpr:
		// 当表达式为二元表达式时，构建相应的SQL表示
		return b.buildBinaryExpr(exp)
	default:
		return errs.NewErrUnsupportedExpressionType(exp)
	}
	return nil
}

// buildSubquery 构建子查询并将其添加到当前查询中
// tab: 子查询对象，包含子查询的构建信息  useAlias: 指示是否使用别名的布尔值。如果为true，则在子查询后添加别名
func (b *builder) buildSubquery(tab Subquery, useAlias bool) error {
	// 调用子查询的Build方法，获取子查询的SQL和参数列表
	q, err := tab.s.Build()
	if err != nil {
		return err
	}
	// 写入左括号
	b.sb.WriteByte('(')
	// 写入子查询的SQL语句，去除最后一个字符（即分号）
	b.sb.WriteString(q.SQL[:len(q.SQL)-1])
	// 如果子查询有参数，将参数添加到当前查询的参数列表中
	if len(q.Args) > 0 {
		b.addArgs(q.Args...)
	}
	// 写入右括号
	b.sb.WriteByte(')')
	// 如果需要使用别名，写入AS关键字和子查询的别名
	b.sb.WriteByte(' ')
	if useAlias {
		b.sb.WriteString(" AS ")
		b.quote(tab.alias)
	}
	return nil
}

// buildBinaryExpr 构建并处理二元表达式。
// 该方法递归地构建二元表达式的左右子表达式，并处理它们之间的操作符。
// 参数 e: 二元表达式对象，包含左子表达式、操作符和右子表达式。
// 返回值 err: 如果在构建过程中发生错误，则返回错误，否则返回nil。
func (b *builder) buildBinaryExpr(e binaryExpr) error {
	err := b.buildSubExpr(e.left)
	if err != nil {
		return err
	}
	if e.op != "" {
		b.sb.WriteByte(' ')
		b.sb.WriteString(e.op.String())
	}
	if e.right != nil {
		b.sb.WriteByte(' ')
		return b.buildSubExpr(e.right)
	}
	return nil
}

// buildSubExpr 处理给定的子表达式，根据其类型构建相应的字符串表示。
// 它支持数学表达式、二元表达式和谓词等不同类型。
// 参数 subExpr: 需要处理的子表达式。
// 返回值: 如果构建过程中发生错误，则返回错误。
func (b *builder) buildSubExpr(subExpr Expression) error {
	switch sub := subExpr.(type) {
	case MathExpr:
		_ = b.sb.WriteByte('(')
		if err := b.buildBinaryExpr(binaryExpr(sub)); err != nil {
			return err
		}
		_ = b.sb.WriteByte(')')
	case binaryExpr:
		_ = b.sb.WriteByte('(')
		if err := b.buildBinaryExpr(sub); err != nil {
			return err
		}
		_ = b.sb.WriteByte(')')
	case Predicate:
		_ = b.sb.WriteByte('(')
		if err := b.buildBinaryExpr(binaryExpr(sub)); err != nil {
			return err
		}
		_ = b.sb.WriteByte(')')
	default:
		if err := b.buildExpression(sub); err != nil {
			return err
		}
	}
	return nil
}

// buildAggregate 构建聚合函数的字符串表示形式。
// 它接受一个 Aggregate 对象(聚合函数)和一个布尔值 useAlias。
func (b *builder) buildAggregate(a Aggregate, useAlias bool) error {
	b.sb.WriteString(a.fn)
	b.sb.WriteByte('(')
	err := b.buildColumn(a.table, a.arg)
	if err != nil {
		return err
	}
	b.sb.WriteByte(')')
	if useAlias {
		b.buildAs(a.alias)
	}
	return nil
}

// buildAs 方法用于在SQL语句中添加别名。
// 如果别名（alias）不为空，则将其添加到构建器（b）中的SQL语句。
// 别名会被适当的引号包围，这是为了在SQL中正确地识别和使用。
func (b *builder) buildAs(alias string) {
	if alias != "" {
		b.sb.WriteString(" AS ")
		b.quote(alias)
	}
}
