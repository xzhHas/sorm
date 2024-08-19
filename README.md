# sorm

> 基于Golang写的orm框架

获取sorm包：

```
go get -u github.com/xzhHas/sorm
```

---

## 查询

sorm.NewSelector[表名] (sorm.DB)

**返回查询结果有两种方法：**

（1）返回单个数据

```go
Get(ctx context.Context) (*T, error)
```

（2）返回多个数据

```go
GetMulti(ctx context.Context) ([]*T, error) 
```

**在使用任何查询前使用以上语法创建一个基础的Selector进行操作。**以下是部分操作例子：



1 **SELECT * FROM `test_model` WHERE `id` = ?**

```go
sorm.NewSelector[TestModel](api.DB).Where(sorm.C("Id").EQ(2)).Get(context.Background())
```

2 **SELECT * FROM `test_model` WHERE (`age` > ?) AND (`age` < ?)**

And有两种使用方法（OR同第一种And方法）：

```go
sorm.NewSelector[TestModel](api.DB).Where(sorm.C("Id").GT(2).And(sorm.C("Id").LT(7))).GetMulti(context.Background())

sorm.NewSelector[TestModel](api.DB).Where(sorm.C("Id").GT(2), sorm.C("Id").LT(7)).GetMulti(context.Background())
```

3 **SELECT `id`,`last_name` FROM `test_model`**

```go
sorm.NewSelector[TestModel](api.DB).Select(sorm.C("Id"), sorm.C("LastName")).GetMulti(context.Background())
```

4 **SELECT `id` AS `my_id`,AVG(`age`) AS `avg_age` FROM `test_model`**

select链式调用，以及AS别名，Raw()自定义字段，Count()，Avg()

```go
sorm.NewSelector[TestModel](api.DB).Select(sorm.C("Id").As("my_id"), sorm.Avg("Age").As("avg_age")).Get(context.Background())
```
