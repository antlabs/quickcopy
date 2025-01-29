package copyslice

// Address represents a nested structure
type Address struct {
	Street  string
	City    string
	Country string
}

// SliceSource contains various types of slices
type SliceSource struct {
	Numbers   []int         // Simple slice type
	Names     []string      // Simple slice type
	Addresses []Address     // Complex slice type
	Matrix    [][]int       // Nested slice type
	Contacts  []ContactInfo // Another complex slice type
}

// ContactInfo represents another nested structure
type ContactInfo struct {
	Email     string
	Phone     string
	Emergency bool
}

// SliceDestination mirrors the source structure
type SliceDestination struct {
	Numbers   []int
	Names     []string
	Addresses []Address
	Matrix    [][]int
	Contacts  []ContactInfo
}

// Person represents a basic person info
type Person struct {
	Name string
	Age  int
}

// Department represents a department with employees
type Department struct {
	Name      string
	Location  string
	Employees []Person
}

// Company represents the source structure with department slice
type Company struct {
	CompanyName string
	Departments []Department
	HeadOffice  string
}

// CompanyDest represents the destination structure
type CompanyDest struct {
	CompanyName string
	Departments []Department
	HeadOffice  string
}

// SourcePerson represents a person in source structure
type SourcePerson struct {
	Name       string
	Age        int
	Department string
	Skills     []string
}

// DestPerson represents a person in destination structure
type DestPerson struct {
	Name       string
	Age        int
	Department string
	Skills     []string
}

// SourceDepartment represents a department in source
type SourceDepartment struct {
	DeptName  string
	Location  string
	Employees []SourcePerson
	Budget    int
	HeadCount int
}

// DestDepartment represents a department in destination
type DestDepartment struct {
	DeptName  string
	Location  string
	Employees []DestPerson
	Budget    int
	HeadCount int
}

// SourceCompany represents the source company structure
type SourceCompany struct {
	CompanyName string
	Departments []SourceDepartment
	HeadOffice  string
	YearFounded int
}

// DestCompany represents the destination company structure
type DestCompany struct {
	CompanyName string
	Departments []DestDepartment
	HeadOffice  string
	YearFounded int
}
