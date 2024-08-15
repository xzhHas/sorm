package test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/gotomicro/ekit"
)

type SimpleStruct struct {
	Id      uint64
	Bool    bool
	BoolPtr *bool

	Int    int
	IntPtr *int

	Int8    int8
	Int8Ptr *int8

	Int16    int16
	Int16Ptr *int16

	Int32    int32
	Int32Ptr *int32

	Int64    int64
	Int64Ptr *int64

	Uint    uint
	UintPtr *uint

	Uint8    uint8
	Uint8Ptr *uint8

	Uint16    uint16
	Uint16Ptr *uint16

	Uint32    uint32
	Uint32Ptr *uint32

	Uint64    uint64
	Uint64Ptr *uint64

	Float32    float32
	Float32Ptr *float32

	Float64    float64
	Float64Ptr *float64

	Byte      byte
	BytePtr   *byte
	ByteArray []byte

	String string

	// 特殊类型
	NullStringPtr *sql.NullString
	NullInt16Ptr  *sql.NullInt16
	NullInt32Ptr  *sql.NullInt32
	NullInt64Ptr  *sql.NullInt64
	NullBoolPtr   *sql.NullBool
	// NullTimePtr    *sql.NullTime
	NullFloat64Ptr *sql.NullFloat64
	JsonColumn     *JsonColumn
}

// JsonColumn 类型定义一个可以存储JSON数据的列。
type JsonColumn struct {
	Val   User // 存储解析后的json数据
	Valid bool // 标识json数据是否有效
}

type User struct {
	Name string
}

// Scan 方法实现将数据库查询结果的一列扫描到 JsonColumn 结构体中。
// src 参数: 需要扫描的源数据，类型为 interface{}，可以是 nil、字符串、字节切片或字节切片的指针。
// 返回值: 如果扫描成功，返回 nil；如果出现错误，返回相应的错误信息。
func (j *JsonColumn) Scan(src any) error {
	// 如果源数据为nil，直接返回nil
	if src == nil {
		return nil
	}
	var bs []byte
	switch val := src.(type) {
	case string:
		bs = []byte(val)
	case []byte:
		bs = val
	case *[]byte:
		if val == nil {
			return nil
		}
		bs = *val
	default:
		return fmt.Errorf("不合法类型 %+v", src)
	}

	if len(bs) == 0 {
		return nil
	}

	err := json.Unmarshal(bs, &j.Val)
	if err != nil {
		return err
	}
	j.Valid = true
	return nil
}

// Value 实现了将 JSON 数据转换为可以存储在数据库中的值。
// 此方法返回与 JSON 数据对应的数据库驱动值，如果数据无效则返回 nil。
// 参数：无
// 返回值：与 JSON 数据对应的数据库驱动值，或在转换失败时返回错误。
func (j *JsonColumn) Value() (driver.Value, error) {
	if !j.Valid {
		return nil, nil
	}
	bs, err := json.Marshal(j.Val)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

// NewSimpleStruct 创建并返回一个新的 SimpleStruct 实例。
// 参数 id 被用作实例的唯一标识符。
func NewSimpleStruct(id uint64) *SimpleStruct {
	return &SimpleStruct{
		Id:             id,
		Bool:           true,
		BoolPtr:        ekit.ToPtr[bool](false),
		Int:            12,
		IntPtr:         ekit.ToPtr[int](13),
		Int8:           8,
		Int8Ptr:        ekit.ToPtr[int8](-8),
		Int16:          16,
		Int16Ptr:       ekit.ToPtr[int16](-16),
		Int32:          32,
		Int32Ptr:       ekit.ToPtr[int32](-32),
		Int64:          64,
		Int64Ptr:       ekit.ToPtr[int64](-64),
		Uint:           14,
		UintPtr:        ekit.ToPtr[uint](15),
		Uint8:          8,
		Uint8Ptr:       ekit.ToPtr[uint8](18),
		Uint16:         16,
		Uint16Ptr:      ekit.ToPtr[uint16](116),
		Uint32:         32,
		Uint32Ptr:      ekit.ToPtr[uint32](132),
		Uint64:         64,
		Uint64Ptr:      ekit.ToPtr[uint64](164),
		Float32:        3.2,
		Float32Ptr:     ekit.ToPtr[float32](-3.2),
		Float64:        6.4,
		Float64Ptr:     ekit.ToPtr[float64](-6.4),
		Byte:           byte(8),
		BytePtr:        ekit.ToPtr[byte](18),
		ByteArray:      []byte("hello"),
		String:         "world",
		NullStringPtr:  &sql.NullString{String: "null string", Valid: true},
		NullInt16Ptr:   &sql.NullInt16{Int16: 16, Valid: true},
		NullInt32Ptr:   &sql.NullInt32{Int32: 32, Valid: true},
		NullInt64Ptr:   &sql.NullInt64{Int64: 64, Valid: true},
		NullBoolPtr:    &sql.NullBool{Bool: true, Valid: true},
		NullFloat64Ptr: &sql.NullFloat64{Float64: 6.4, Valid: true},
		JsonColumn: &JsonColumn{
			Val:   User{Name: "Tom"},
			Valid: true,
		},
	}
}
