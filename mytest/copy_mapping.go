package mytest

type mappingSource struct {
	ModuleName string
	FieldName2 int
}

type mappingDestination struct {
	Module string
	Field2 int
}

// :quickcopy Module = ModuleName, Field2 = FieldName2
func mappingCopy(dst *mappingDestination, src *mappingSource) {
	dst.Field2 =
		src.
			FieldName2
	dst.Module = src.ModuleName
}

type Source struct {
	SubStruct struct {
		FieldName string
	}
}

type Destination struct {
	SubStruct struct {
		Field string
	}
}

// TODO:quickcopy SubStruct.Field = SubStruct.FieldName
func Copy(dst *Destination, src *Source) {
}
