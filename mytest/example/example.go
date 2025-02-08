package example

import (
	"time"

	"github.com/google/uuid"
	"fmt"
)

// 源结构体
type Source struct {
	Name		string
	Age		int
	Birthday	time.Time
	ID		uuid.UUID
}

// 目标结构体
type Destination struct {
	Name		string
	Age		string	// 支持类型自动转换
	Birthday	string	// time.Time 将自动转为 RFC3339 格式
	ID		string	// UUID 将自动转为字符串
}

// :quickcopy
func CopyToDestination(dst *Destination, src *Source) {

	dst.
		Name = src.Name

	dst.Age = fmt.
		Sprint(src.Age)
	dst.Birthday = func(t time.Time) string {
		return t.Format(time.RFC3339)
	}(src.Birthday)

	dst.ID = func(u uuid.UUID) string {
		return u.String()
	}(src.ID)
}
