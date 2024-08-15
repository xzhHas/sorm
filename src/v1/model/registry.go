package model

import (
	"reflect"
	"sorm/src/v1/internal/errs"
	"strings"
	"sync"
	"unicode"
)

// Option 元数据注册中心的配置项，用于修改Model实例
type Option func(m *Model) error

// Registry 元数据注册中心的抽象（接口）
type Registry interface {
	// Get 查找并返回已注册的元数据模型
	Get(val any) (*Model, error)
	// Register 注册一个新模型，并返回创建的元数据模型
	Register(val any, opts ...Option) (*Model, error)
}

// registry 基于标签和接口的实现
// 目前来看，我们只有一个实现，所以暂时可以维持私有
// 使用同步映射`sync.Map`来存储注册过的模型（并发安全）
type registry struct {
	models sync.Map
	/*  这两个接口均已实现，及registry这个结构体实现了Registry接口
	Get(val any)(*Model,error)
	Register(val any,opts ...Option)(*Model,error)
	尽管有额外的方法，到时已经实现了Registry接口
	*/
}

// NewRegistry 创建并返回一个新的元数据注册中心实例
func NewRegistry() Registry {
	return &registry{}
}

// Get 查找并返回已注册的元数据模型，如果模型未注册，则尝试注册后返回
func (r *registry) Get(val any) (*Model, error) {
	typ := reflect.TypeOf(val)
	m, ok := r.models.Load(typ)
	if ok {
		return m.(*Model), nil
	}
	return r.Register(val)
}

// Register 注册一个新的模型，并返回创建的元数据模型
func (r *registry) Register(val any, opts ...Option) (*Model, error) {
	m, err := r.parseModel(val)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		err = opt(m)
		if err != nil {
			return nil, err
		}
	}
	typ := reflect.TypeOf(val)
	r.models.Store(typ, m)
	return m, nil
}

// parseModel 支持从标签中提取自定义设置
// 标签形式 orm:"key1=value1,key2=value2"
// 解析模型结构体，并创建元数据模型
func (r *registry) parseModel(val any) (*Model, error) {
	typ := reflect.TypeOf(val)
	if typ.Kind() != reflect.Ptr ||
		typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	typ = typ.Elem()

	// 获得字段的数量
	numField := typ.NumField()
	fieldMap := make(map[string]*Field, numField)
	fields := make([]*Field, 0, numField)
	colMap := make(map[string]*Field, numField)
	for i := 0; i < numField; i++ {
		fdType := typ.Field(i)
		tags, err := r.parseTag(fdType.Tag)
		if err != nil {
			return nil, err
		}
		colName := tags[tagKeyColumn]
		if colName == "" {
			colName = underscoreName(fdType.Name)
		}
		f := &Field{
			ColName: colName,
			Type:    fdType.Type,
			GoName:  fdType.Name,
			Offset:  fdType.Offset,
			Index:   i,
		}
		fieldMap[fdType.Name] = f
		fields = append(fields, f)
		colMap[colName] = f
	}
	var tableName string
	if tn, ok := val.(TableName); ok {
		tableName = tn.TableName()
	}

	if tableName == "" {
		tableName = underscoreName(typ.Name())
	}

	return &Model{
		TableName: tableName,
		FieldMap:  fieldMap,
		ColumnMap: colMap,
		Fields:    fields,
	}, nil
}

// parseTag 解析结构体标签中的 `orm` 标签内容
func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag := tag.Get("orm")
	if ormTag == "" {
		// 返回一个空的 map，这样调用者就不需要判断 nil 了
		return map[string]string{}, nil
	}
	// 这个初始化容量就是我们支持的 key 的数量，
	// 现在只有一个，所以我们初始化为 1
	res := make(map[string]string, 1)

	// 接下来就是字符串处理了
	pairs := strings.Split(ormTag, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, errs.NewErrInvalidTagContent(pair)
		}
		res[kv[0]] = kv[1]
	}
	return res, nil
}

// underscoreName 将驼峰命名转换为下划线命名
func underscoreName(tableName string) string {
	var buf []byte
	for i, v := range tableName {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}

	}
	return string(buf)
}

// WithTableName 用于设置模型的表名
func WithTableName(tableName string) Option {
	return func(model *Model) error {
		model.TableName = tableName
		return nil
	}
}

// WithColumnName 用于设置模型字段的数据库列名
func WithColumnName(field string, columnName string) Option {
	return func(model *Model) error {
		fd, ok := model.FieldMap[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}
		// 注意，这里我们根本没有检测 ColName 会不会是空字符串
		// 因为正常情况下，用户都不会写错
		// 即便写错了，也很容易在测试中发现
		fd.ColName = columnName
		return nil
	}
}
