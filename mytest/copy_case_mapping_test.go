package mytest

import "testing"

// 测试忽略大小写的字段映射
func TestCopyCaseTest(t *testing.T) {
	src := &CaseSource{
		Username: "张三",
		UserAge:  25,
	}
	dst := &CaseDestination{}

	CaseCopy(dst, src)

	if dst.username != src.Username {
		t.Errorf("username: 期望 %s, 得到 %s", src.Username, dst.username)
	}
	if dst.userage != src.UserAge {
		t.Errorf("userage: 期望 %d, 得到 %d", src.UserAge, dst.userage)
	}
}

// 测试规则字段映射
func TestCopyRuleTest(t *testing.T) {
	src := &RuleSource{
		Age: 18,
	}
	dst := &RuleDestination{}

	RuleCopy(dst, src)

	if dst.UserAge != src.Age {
		t.Errorf("UserAge: 期望 %d, 得到 %d", src.Age, dst.UserAge)
	}
}
