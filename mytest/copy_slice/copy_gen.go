package copyslice

func CompanySliceCopy(dst *CompanyDest, src *Company) {

	dst.CompanyName = src.CompanyName

	dst.Departments = src.Departments

	dst.HeadOffice = src.HeadOffice
}
func SliceCopy(dst *SliceDestination, src *SliceSource) {
	dst.
		Numbers = src.Numbers
	dst.Names = src.Names
	dst.Addresses = src.Addresses
	dst.Matrix = src.Matrix
	dst.Contacts = src.Contacts
}

func DiffTypeSliceCopy(dst *DestCompany, src *SourceCompany) {

	dst.
		CompanyName = src.CompanyName
	dst.Departments = copySliceDestDepartmentFromSliceSourceDepartment(src.Departments)
	dst.HeadOffice = src.HeadOffice
	dst.YearFounded = src.YearFounded
}
func copySourcePersonFromSourcePerson(dst *SourcePerson, src *SourcePerson) {

	dst.Name = src.Name
	dst.Age = src.Age
	dst.Department = src.Department
	dst.Skills = src.Skills
}
func copyDestPersonFromDestPerson(dst *DestPerson, src *DestPerson) {

	dst.Name = src.Name
	dst.
		Age = src.Age
	dst.Department = src.Department
	dst.Skills = src.Skills
}
func copyDestPersonFromSourcePerson(dst *DestPerson, src *SourcePerson) {

	dst.Name = src.Name
	dst.Age = src.Age
	dst.Department = src.Department
	dst.Skills = src.Skills
}

func copySliceDestPersonFromSliceSourcePerson(src []SourcePerson) []DestPerson {
	if src == nil {
		return nil
	}
	dst := make([]DestPerson, len(src))
	for i := range src {
		copyDestPersonFromSourcePerson(&dst[i], &src[i])
	}
	return dst
}
func copyDestDepartmentFromSourceDepartment(dst *DestDepartment, src *SourceDepartment) {
	dst.DeptName = src.DeptName
	dst.Location = src.Location
	dst.Employees = copySliceDestPersonFromSliceSourcePerson(src.Employees)
	dst.Budget = src.Budget
	dst.HeadCount = src.HeadCount
}

func copySliceDestDepartmentFromSliceSourceDepartment(src []SourceDepartment) []DestDepartment {
	if src == nil {
		return nil
	}
	dst := make([]DestDepartment, len(src))
	for i := range src {
		copyDestDepartmentFromSourceDepartment(&dst[i], &src[i])
	}
	return dst
}
