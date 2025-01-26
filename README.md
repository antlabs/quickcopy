# quickcopy

基于静态代码生成的深度拷贝函数生成工具，支持基础类型、时间和 UUID 等类型的自动转换。

## 特性

- 🚀 自动生成结构体间的深度拷贝函数
- 💪 支持多种类型转换：
  - int 和 string 互转
  - time.Time 和 string 互转
  - uuid.UUID 和 string 互转
  - int 和 int8/16/32/64 互转
- 🎯 使用简单，仅需一行注释即可生成
- ⚡ 基于静态代码生成，运行时零开销

## 功能详情

- **类型转换逻辑**：
  - 支持整数类型宽度判断，确保类型转换的安全性。
  - 提供类型转换逻辑获取功能，自动处理不同类型间的转换。

- **窄化转换**：
  - 默认不允许窄化转换，可通过 `--allow-narrow` 选项进行配置。
  - `--allow-narrow` 选项允许在类型转换时进行窄化，例如从 int64 到 int32。

- **忽略大小写**：
  - 默认情况下，字段名比较区分大小写。
  - 可以通过 `--ignore-case` 选项来忽略字段名的大小写，使得字段名的匹配不区分大小写。

- **模糊字段映射**：
  - 支持通过注释指定源结构体和目标结构体之间的字段映射规则。
  - 可以处理字段名称不完全匹配的情况。
  - 未指定映射规则的字段会自动按名称匹配（支持忽略大小写）。
  - 支持嵌套结构体的字段映射

- **单个元素与数组转换**：
  - 支持将单个元素赋值给数组，以及从数组中提取单个元素进行赋值。
  - 可以通过 `--single-to-slice` 选项启用此功能。
  - 支持以下转换场景：
    - 单个元素转换为长度为1的数组
    - 从数组中提取第一个元素赋值给单个字段
    - 自动处理类型转换，如 `int` 到 `[]string`

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

## 使用示例

以下是一个使用 quickcopy 生成拷贝函数的示例：

```go
// 自动生成的拷贝函数示例
func CopyData(dst *Destination, src *Source) {
    dst.Name = src.Name
    dst.Age = strconv.Itoa(src.Age) // int 转 string
    dst.Birthday = src.Birthday.Format(time.RFC3339) // time.Time 转 string
    dst.ID = src.ID.String() // UUID 转 string
}
```

## 配置选项

### `--ignore-case`
例如：
```go
// 源结构体
type Source struct {
    UserName string
    Age      int
}

// 目标结构体
type Destination struct {
    username string  // 字段名大小写不同
    Age      string
}
```

添加拷贝函数原型和注释标记：

```go
// :quickcopy --ignore-case
func CopyToDestination(dst *Destination, src *Source) {
}
```
运行
```bash
quickcopy
```

将生成如下拷贝函数：
```go
func CopyToDestination(dst *Destination, src *Source) {
    dst.username = src.UserName  // 自动匹配不同大小写的字段
    dst.Age = fmt.Sprint(src.Age)
}
```

### `--field-mapping`
例如：
```go
// 源结构体
type Source struct {
    FirstName string
    LastName  string
    UserID    int
}

// 目标结构体
type Destination struct {
    FullName string  // 需要合并 FirstName 和 LastName
    ID       string  // 映射自 UserID
}
```

添加拷贝函数原型和注释标记，并指定字段映射规则：

```go
// :quickcopy
// FullName=FirstName
// ID=UserID
func CopyToDestination(dst *Destination, src *Source) {
}
```

将生成如下拷贝函数：
```go
func CopyToDestination(dst *Destination, src *Source) {
    dst.FullName = src.FirstName // 按照映射规则合并字段
    dst.ID = fmt.Sprint(src.UserID)                    // 按照映射规则转换字段
}
```

### `--single-to-slice`
例如：
```go
// 源结构体
type Source struct {
    Tag    string   // 单个标签
    Status int      // 单个状态
    IDs    []int    // ID列表
}

// 目标结构体
type Destination struct {
    Tags    []string // 标签列表
    Statuses []string // 状态列表
    ID      int      // 单个ID
}
```

添加拷贝函数原型和注释标记：

```go
// :quickcopy --single-to-slice
func CopyToDestination(dst *Destination, src *Source) {
}
```

将生成如下拷贝函数：
```go
func CopyToDestination(dst *Destination, src *Source) {
    dst.Tags = []string{src.Tag}           // 单个字符串转换为字符串数组
    dst.Statuses = []string{fmt.Sprint(src.Status)} // 单个int转换为字符串数组
    if len(src.IDs) > 0 {
        dst.ID = src.IDs[0]                // 从数组中提取第一个元素
    }
}
```

这个例子展示了三种常见的转换场景：
1. 将单个字符串 `Tag` 转换为字符串数组 `Tags`
2. 将单个整数 `Status` 转换为字符串数组 `Statuses`（包含类型转换）
3. 从整数数组 `IDs` 中提取第一个元素赋值给单个整数 `ID`

## 支持的类型转换

| 源类型 | 目标类型 | 转换方式 |
|--------|----------|----------|
| `int`          | `string`       | fmt.Sprint                        |
| `string`       | `int`          | strconv.Atoi                     |
| `time.Time`    | `string`       | Format(time.RFC3339)             |
| `string`       | `time.Time`    | Parse(time.RFC3339)              |
| `uuid.UUID`    | `string`       | UUID.String()                   |
| `string`       | `uuid.UUID`    | uuid.Parse                      |
| `int8`         | `int16`        | int16(i)                        |
| `int8`         | `int32`        | int32(i)                        |
| `int8`         | `int64`        | int64(i)                        |
| `int16`        | `int8`         | int8(i)                         |
| `int16`        | `int32`        | int32(i)                        |
| `int16`        | `int64`        | int64(i)                        |
| `int32`        | `int8`         | int8(i)                         |
| `int32`        | `int16`        | int16(i)                        |
| `int32`        | `int64`        | int64(i)                        |
| `int64`        | `int8`         | int8(i)                         |
| `int64`        | `int16`        | int16(i)                        |
| `int64`        | `int32`        | int32(i)                        |
| `[]int`        | `int`          | 数组的第一个元素赋值给单个整数     |
| `int`          | `[]int`        | 单个整数赋值给数组（如果启用）     |

## 代码结构概览

- **`CopyFuncInfo` 结构体**：存储拷贝函数的信息，包括源和目标变量、类型及字段映射。
- **`FieldMapping` 结构体**：定义字段间的映射关系和转换逻辑。
- **主要功能函数**：
  - `generateCompleteCopyFunc`：生成完整的拷贝函数并替换原始函数。
  - `getFieldMappings`：获取字段映射关系。

## 自定义扩展

可以通过修改以下函数来自定义转换逻辑：

1. `getTypeConversion`: 自定义类型转换逻辑

## 注意事项

1. 对于 string 到 int 等可能失败的转换，生成的代码会静默处理错误
2. 时间类型统一使用 RFC3339 格式进行转换
3. 确保目标结构体字段类型在支持的转换范围内

## 常见问题和解决方案

- 确保所有依赖项已正确安装。
- 验证 Go 环境设置是否正确。
- 参考 GitHub 问题页面以获取已知问题和解决方案。