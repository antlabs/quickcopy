package quickcopy

func isBasicType(typeName string) bool {
	switch typeName {
	case "string", "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "bool":
		return true
	default:
		return false
	}
}

func isIntegerType(typeName string) bool {
	switch typeName {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
		return true
	default:
		return false
	}
}

// getIntWidth 获取整数类型的宽度
func getIntWidth(typeName string) int {
	switch typeName {
	case "int8", "uint8":
		return 8
	case "int16", "uint16":
		return 16
	case "int32", "uint32":
		return 32
	case "int64", "uint64":
		return 64
	case "int", "uint":
		// 假设 int 和 uint 是 64 位的
		return 64
	case "float32":
		return 32
	case "float64":
		return 64
	default:
		return 0
	}
}

// TOOD 加unsafe开关
func handleSpecialTypeConversion(srcType, dstType string) string {
	switch {
	case srcType == "string" && dstType == "float64":
		return "func(s string) float64 { f, _ := strconv.ParseFloat(s, 64); return f }"
	case srcType == "float64" && dstType == "string":
		return "func(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }"
	case srcType == "string" && dstType == "[]byte":
		return "func(s string) []byte { return []byte(s) }"
	case srcType == "[]byte" && dstType == "string":
		return "func(b []byte) string { return string(b) }"
	case srcType == "int" && dstType == "string":
		return "fmt.Sprint"
	case srcType == "string" && dstType == "int":
		return "func(s string) int { i, _ := strconv.Atoi(s); return i }"
	case srcType == "time.Time" && dstType == "string":
		return "func(t time.Time) string { return t.Format(time.RFC3339) }"
	case srcType == "string" && dstType == "time.Time":
		return "func(s string) time.Time { t, _ := time.Parse(time.RFC3339, s); return t }"
	case srcType == "uuid.UUID" && dstType == "string":
		return "func(u uuid.UUID) string { return u.String() }"
	case srcType == "string" && dstType == "uuid.UUID":
		return "func(s string) uuid.UUID { u, _ := uuid.Parse(s); return u }"
	}
	return ""
}
