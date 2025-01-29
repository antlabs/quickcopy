package copyslice

import "testing"

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
