# quickcopy

基于静态代码生成的深度拷贝函数生成工具，支持基础类型、时间和 UUID 等类型的自动转换。

## 特性

- 🚀 自动生成结构体间的深度拷贝函数
- 💪 支持多种类型转换：
  - int 和 string 互转
  - time.Time 和 string 互转
  - uuid.UUID 和 string 互转
- 🎯 使用简单，仅需一行注释即可生成
- ⚡ 基于静态代码生成，运行时零开销

## 安装

### 方式1：直接安装

```bash
go get github.com/antlabs/quickcopy/cmd/quickcopy
```

### 方式2：从源码编译

```bash
git clone https://github.com/antlabs/quickcopy.git
cd quickcopy
make
```

## 使用方法

### 步骤1：定义结构体

在你的代码中定义源结构体和目标结构体：

```go
// 源结构体
type Source struct {
    Name     string
    Age      int
    Birthday time.Time
    ID       uuid.UUID
}

// 目标结构体
type Destination struct {
    Name     string
    Age      string    // 支持类型自动转换
    Birthday string    // time.Time 将自动转为 RFC3339 格式
    ID       string    // UUID 将自动转为字符串
}
```

### 步骤2：添加拷贝函数原型和注释标记

添加拷贝函数原型和注释标记：

```go
// :quickcopy
func CopyToDestination(dst *Destination, src *Source) {
}
```

### 步骤3：运行工具生成代码

运行工具生成代码：

```bash
quickcopy
```

工具会自动生成如下拷贝函数：

```go
// CopyToDestination 是一个自动生成的拷贝函数
func CopyToDestination(dst *Destination, src *Source) {
    dst.Name = src.Name
    dst.Age = fmt.Sprint(src.Age)
    dst.Birthday = src.Birthday.Format(time.RFC3339)
    dst.ID = src.ID.String()
}
```

## 支持的类型转换

| 源类型 | 目标类型 | 转换方式 |
|--------|----------|----------|
| int | string | fmt.Sprint |
| string | int | strconv.Atoi |
| time.Time | string | Format(time.RFC3339) |
| string | time.Time | Parse(time.RFC3339) |
| uuid.UUID | string | UUID.String() |
| string | uuid.UUID | uuid.Parse |

## 自定义扩展

可以通过修改以下函数来自定义转换逻辑：

1. `FieldMapping`: 自定义字段映射关系
2. `getTypeConversion`: 自定义类型转换逻辑

## 注意事项

1. 对于 string 到 int 等可能失败的转换，生成的代码会静默处理错误
2. 时间类型统一使用 RFC3339 格式进行转换
3. 确保目标结构体字段类型在支持的转换范围内