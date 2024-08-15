package model

import "reflect"

// Model 代表一个数据库表的模型，包含了表名及所有字段的描述信息
type Model struct {
	// TableName 结构体对应的表名
	TableName string
	// Fields 指针切片（切片里面放的全是指针），包含了该模型所有的字段信息
	Fields []*Field
	// FieldMap 是一个由字段名映射到Field结构体的map，便于快速查找字段
	FieldMap map[string]*Field
	// ColumnMap 是一个由数据库列名映射到Field结构体的map，便于快速查找字段
	ColumnMap map[string]*Field
}

// Field 结构体，描述一个字段与其对应的数据库列之间的映射关系
type Field struct {
	// ColName 对应数据库表的列名
	ColName string
	// GoName 对应 Go 结构体字段名
	GoName string
	// Type 字段的反射类型，用于了解字段的Go类型
	Type reflect.Type
	// Index 字段在结构体中的索引
	Index int
	// Offset 相对于对象起始地址的偏移量，单位为字节，用于反射操作时定位字段内存位置
	Offset uintptr
}

// tagKeyColumn 是一个常量，定义了Go结构体标签中用于指定数据库列名的键值
// 我们支持的全部标签上的 key 都放在这里
// 方便用户查找，和我们后期维护
const (
	tagKeyColumn = "column"
)

// 用户自定义一些模型信息的接口，集中放在这里
// 方便用户查找和我们后期维护

// TableName 用户实现这个接口来返回自定义的表名，可以自定义对应结构体的表名
type TableName interface {
	TableName() string
}
