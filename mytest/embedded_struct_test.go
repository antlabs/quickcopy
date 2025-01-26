package mytest

import (
	"fmt"
	"testing"
	"time"
)

// 基础用户信息
type UserBase struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// 源结构体：完整的用户信息
type UserSource struct {
	UserBase // 嵌入基础用户信息
	Name     string
	Age      int
	Email    string
}

// 目标结构体：用户视图
type UserView struct {
	UserBase // 嵌入基础用户信息
	Name     string
	Age      string // 类型转换：int -> string
	Contact  string // 映射自 Email
}

// :quickcopy Contact=Email
func CopyToUserView(dst *UserView, src *UserSource) {

	dst.Contact = src.Email
	dst.ID = src.ID
	dst.CreatedAt = src.
		CreatedAt

	dst.
		UpdatedAt = src.UpdatedAt

	dst.Name = src.Name
	dst.
		Age = fmt.Sprint(src.Age)
}

func TestEmbeddedStructCopy(t *testing.T) {
	src := &UserSource{
		UserBase: UserBase{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:  "张三",
		Age:   25,
		Email: "zhangsan@example.com",
	}

	dst := &UserView{}
	CopyToUserView(dst, src)

	// 验证嵌入字段的复制
	if src.ID != dst.ID {
		t.Errorf("ID not copied correctly, got %d, want %d", dst.ID, src.ID)
	}
	if !src.CreatedAt.Equal(dst.CreatedAt) {
		t.Errorf("CreatedAt not copied correctly")
	}
	if !src.UpdatedAt.Equal(dst.UpdatedAt) {
		t.Errorf("UpdatedAt not copied correctly")
	}

	// 验证普通字段的复制
	if src.Name != dst.Name {
		t.Errorf("Name not copied correctly")
	}
	if dst.Age != "25" {
		t.Errorf("Age not converted correctly")
	}
	if dst.Contact != src.Email {
		t.Errorf("Email not mapped correctly to Contact")
	}
}
