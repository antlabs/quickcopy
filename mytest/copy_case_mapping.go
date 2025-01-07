package mytest

// Source struct with fields in one case
type CaseSource struct {
	Username    string
	UserAge     int
	UserAddress string
}

// Destination struct with fields in different case
type CaseDestination struct {
	username    string
	userage     int
	useraddress string
}

// Test case-insensitive field mapping
// :quickcopy --ignore-case
func CaseCopy(dst *CaseDestination, src *CaseSource) {
	dst.username = src.Username

	dst.userage = src.UserAge
	dst.useraddress = src.
		UserAddress
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
