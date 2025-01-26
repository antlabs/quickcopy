package mytest

import "testing"

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

// :quickcopy
func SliceCopy(dst *SliceDestination, src *SliceSource) {

	dst.Numbers = src.Numbers

	dst.Names = src.Names
	dst.Addresses = src.Addresses
	dst.Matrix = src.Matrix
	dst.Contacts = src.Contacts
}

func TestCopySliceTest(t *testing.T) {
	src := &SliceSource{
		Numbers: []int{1, 2, 3, 4, 5},
		Names:   []string{"Alice", "Bob", "Charlie"},
		Addresses: []Address{
			{Street: "123 Main St", City: "New York", Country: "USA"},
			{Street: "456 Park Ave", City: "London", Country: "UK"},
		},
		Matrix: [][]int{
			{1, 2, 3},
			{4, 5, 6},
		},
		Contacts: []ContactInfo{
			{Email: "alice@example.com", Phone: "123-456", Emergency: true},
			{Email: "bob@example.com", Phone: "789-012", Emergency: false},
		},
	}

	dst := &SliceDestination{}
	SliceCopy(dst, src)

	// Test simple slice types
	if len(dst.Numbers) != len(src.Numbers) {
		t.Errorf("Numbers slice length mismatch")
	}
	if len(dst.Names) != len(src.Names) {
		t.Errorf("Names slice length mismatch")
	}

	// Test complex slice types
	if len(dst.Addresses) != len(src.Addresses) {
		t.Errorf("Addresses slice length mismatch")
	}
	if len(dst.Matrix) != len(src.Matrix) {
		t.Errorf("Matrix slice length mismatch")
	}
	if len(dst.Contacts) != len(src.Contacts) {
		t.Errorf("Contacts slice length mismatch")
	}

	// Test content of complex types
	for i, addr := range src.Addresses {
		if dst.Addresses[i].Street != addr.Street ||
			dst.Addresses[i].City != addr.City ||
			dst.Addresses[i].Country != addr.Country {
			t.Errorf("Address at index %d does not match", i)
		}
	}

	// Test content of contact info
	for i, contact := range src.Contacts {
		if dst.Contacts[i].Email != contact.Email ||
			dst.Contacts[i].Phone != contact.Phone ||
			dst.Contacts[i].Emergency != contact.Emergency {
			t.Errorf("Contact at index %d does not match", i)
		}
	}
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

// :quickcopy
func CompanySliceCopy(dst *CompanyDest, src *Company) {

	dst.CompanyName = src.CompanyName

	dst.Departments = src.Departments

	dst.HeadOffice = src.HeadOffice
}

func TestCompanySliceCopy(t *testing.T) {
	src := &Company{
		CompanyName: "Tech Corp",
		Departments: []Department{
			{
				Name:     "Engineering",
				Location: "Floor 1",
				Employees: []Person{
					{Name: "Alice", Age: 30},
					{Name: "Bob", Age: 25},
				},
			},
			{
				Name:     "Marketing",
				Location: "Floor 2",
				Employees: []Person{
					{Name: "Charlie", Age: 28},
					{Name: "David", Age: 32},
				},
			},
		},
		HeadOffice: "New York",
	}

	dst := &CompanyDest{}
	CompanySliceCopy(dst, src)

	// Test company name
	if dst.CompanyName != src.CompanyName {
		t.Errorf("CompanyName mismatch, got %v, want %v", dst.CompanyName, src.CompanyName)
	}

	// Test departments length
	if len(dst.Departments) != len(src.Departments) {
		t.Errorf("Departments length mismatch, got %d, want %d", len(dst.Departments), len(src.Departments))
	}

	// Test each department and its employees
	for i, dept := range src.Departments {
		if dst.Departments[i].Name != dept.Name {
			t.Errorf("Department name mismatch at index %d", i)
		}
		if dst.Departments[i].Location != dept.Location {
			t.Errorf("Department location mismatch at index %d", i)
		}

		// Test employees in each department
		if len(dst.Departments[i].Employees) != len(dept.Employees) {
			t.Errorf("Employees length mismatch in department %d", i)
		}

		for j, emp := range dept.Employees {
			if dst.Departments[i].Employees[j].Name != emp.Name {
				t.Errorf("Employee name mismatch at department %d, employee %d", i, j)
			}
			if dst.Departments[i].Employees[j].Age != emp.Age {
				t.Errorf("Employee age mismatch at department %d, employee %d", i, j)
			}
		}
	}

	// Test head office
	if dst.HeadOffice != src.HeadOffice {
		t.Errorf("HeadOffice mismatch, got %v, want %v", dst.HeadOffice, src.HeadOffice)
	}
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

func TestDiffTypeSliceCopy(t *testing.T) {
	src := &SourceCompany{
		CompanyName: "Tech Corp",
		HeadOffice:  "New York",
		YearFounded: 2000,
		Departments: []SourceDepartment{
			{
				DeptName:  "Engineering",
				Location:  "Floor 1",
				Budget:    1000000,
				HeadCount: 50,
				Employees: []SourcePerson{
					{
						Name:       "Alice",
						Age:        30,
						Department: "Engineering",
						Skills:     []string{"Go", "Python", "Docker"},
					},
					{
						Name:       "Bob",
						Age:        25,
						Department: "Engineering",
						Skills:     []string{"Java", "Kubernetes"},
					},
				},
			},
			{
				DeptName:  "Marketing",
				Location:  "Floor 2",
				Budget:    500000,
				HeadCount: 30,
				Employees: []SourcePerson{
					{
						Name:       "Charlie",
						Age:        28,
						Department: "Marketing",
						Skills:     []string{"SEO", "Content Writing"},
					},
					{
						Name:       "David",
						Age:        32,
						Department: "Marketing",
						Skills:     []string{"Social Media", "Analytics"},
					},
				},
			},
		},
	}

	dst := &DestCompany{}
	DiffTypeSliceCopy(dst, src)

	// Test basic company info
	if dst.CompanyName != src.CompanyName {
		t.Errorf("CompanyName mismatch, got %v, want %v", dst.CompanyName, src.CompanyName)
	}
	if dst.YearFounded != src.YearFounded {
		t.Errorf("YearFounded mismatch, got %v, want %v", dst.YearFounded, src.YearFounded)
	}

	// Test departments
	if len(dst.Departments) != len(src.Departments) {
		t.Errorf("Departments length mismatch, got %d, want %d", len(dst.Departments), len(src.Departments))
	}

	for i, srcDept := range src.Departments {
		dstDept := dst.Departments[i]
		if dstDept.DeptName != srcDept.DeptName {
			t.Errorf("Department name mismatch at index %d", i)
		}
		if dstDept.Budget != srcDept.Budget {
			t.Errorf("Department budget mismatch at index %d", i)
		}
		if dstDept.HeadCount != srcDept.HeadCount {
			t.Errorf("Department headcount mismatch at index %d", i)
		}

		// Test employees
		if len(dstDept.Employees) != len(srcDept.Employees) {
			t.Errorf("Employees length mismatch in department %d", i)
		}

		for j, srcEmp := range srcDept.Employees {
			dstEmp := dstDept.Employees[j]
			if dstEmp.Name != srcEmp.Name {
				t.Errorf("Employee name mismatch at department %d, employee %d", i, j)
			}
			if dstEmp.Age != srcEmp.Age {
				t.Errorf("Employee age mismatch at department %d, employee %d", i, j)
			}
			if dstEmp.Department != srcEmp.Department {
				t.Errorf("Employee department mismatch at department %d, employee %d", i, j)
			}

			// Test skills slice
			if len(dstEmp.Skills) != len(srcEmp.Skills) {
				t.Errorf("Skills length mismatch for employee at department %d, employee %d", i, j)
			}
			for k, skill := range srcEmp.Skills {
				if dstEmp.Skills[k] != skill {
					t.Errorf("Skill mismatch at department %d, employee %d, skill %d", i, j, k)
				}
			}
		}
	}
}

// :quickcopy
func DiffTypeSliceCopy(dst *DestCompany, src *SourceCompany) {

}
