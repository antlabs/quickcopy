package quickcopy

import "strings"

// 工具函数
func sanitizeTypeName(typeName string) string {
	// 保留包名前缀但替换非法字符
	typeName = strings.ReplaceAll(typeName, ".", "_")
	typeName = strings.ReplaceAll(typeName, "[]", "Slice_")
	typeName = strings.ReplaceAll(typeName, "*", "Ptr_")
	typeName = strings.ReplaceAll(typeName, "[", "Arr_")
	typeName = strings.ReplaceAll(typeName, "]", "")

	// 首字母大写
	if len(typeName) > 0 {
		typeName = strings.ToUpper(typeName[:1]) + typeName[1:]
	}
	return typeName
}

// extractFieldName 从字段路径中提取字段名
func extractFieldName(fieldPath string) string {
	// 按点号分割路径
	parts := strings.Split(fieldPath, ".")
	if len(parts) == 0 {
		return ""
	}
	// 返回最后一个部分（字段名）
	return parts[len(parts)-1]
}
