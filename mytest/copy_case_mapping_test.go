package mytest

import "testing"

// Source struct with fields in one case
type CaseSource struct {
	Username	string
	UserAge		int
	UserAddress	string
}

// Destination struct with fields in different case
type CaseDestination struct {
	username	string
	userage		int
	useraddress	string
}

// Test case-insensitive field mapping
// :quickcopy --ignore-case
func CaseCopy(dst *CaseDestination, src *CaseSource) {

	dst.username = src.Username
	dst.userage =
		src.UserAge

	dst.useraddress = src.UserAddress
}

// Source struct for rule-based mapping
type RuleSource struct {
	Age int
}

// Destination struct for rule-based mapping
type RuleDestination struct {
	UserAge int
}

// Test rule-based field mapping
//
// :quickcopy UserAge = Age
func RuleCopy(dst *RuleDestination, src *RuleSource) {

	dst.UserAge = src.Age
}

// 测试忽略大小写的字段映射
func TestCopyCaseTest(t *testing.T) {
	src := &CaseSource{
		Username:	"张三",
		UserAge:	25,
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
