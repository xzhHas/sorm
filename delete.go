package sorm

import (
	"fmt"
	"reflect"
	"strings"
)

// Deleter 结构体表示一个用于执行数据库删除操作的对象
// 参数 T 表示要删除的数据类型
type Deleter[T any] struct {
	sb    strings.Builder
	where []Predicate
	table string
	args  []any
	sess  session
}

func NewDeleter[T any](sess session) *Deleter[T] {
	return &Deleter[T]{}
}

func (d *Deleter[T]) Build() (*Query, error) {
	d.sb.WriteString("DELETE FROM ")
	if d.table == "" {
		var t T
		d.sb.WriteByte('`')
		d.sb.WriteString(reflect.TypeOf(t).Name())
		d.sb.WriteByte('`')
	} else {
		d.sb.WriteString(d.table)
	}

	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")
		for i, p := range d.where {
			if i > 0 {
				d.sb.WriteString(" AND ")
			}
			if err := d.buildExpression(p); err != nil {
				return nil, err
			}
		}
	}
	d.sb.WriteString(";")
	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, nil
}

func (d *Deleter[T]) buildExpression(e Expression) error {
	if e == nil {
		return nil
	}
	switch exp := e.(type) {
	case Column:
		d.sb.WriteByte('`')
		d.sb.WriteString(exp.name)
		d.sb.WriteByte('`')
	case value:
		d.sb.WriteByte('?')
		d.args = append(d.args, exp.val)
	case Predicate:
		// 处理左侧表达式
		if err := d.buildExpression(exp.left); err != nil {
			return err
		}
		// 处理操作符
		d.sb.WriteString(" ")
		d.sb.WriteString(string(exp.op))
		d.sb.WriteString(" ")
		// 处理右侧表达式
		if err := d.buildExpression(exp.right); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported expression type: %T", e)
	}
	return nil
}

// From 方法用于设置要操作的表名或表引用
func (d *Deleter[T]) From(tbl string) *Deleter[T] {
	d.table = tbl
	return d
}

// Where 方法用于设置删除操作的条件
func (d *Deleter[T]) Where(ps ...Predicate) *Deleter[T] {
	d.where = ps
	return d
}
