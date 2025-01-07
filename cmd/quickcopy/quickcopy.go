package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

// CopyFuncInfo 存储拷贝函数信息
type CopyFuncInfo struct {
	FuncName string
	SrcVar   string
	DstVar   string
	SrcType  string
	DstType  string
	Fields   []FieldMapping
}

// FieldMapping 存储字段映射关系
type FieldMapping struct {
	SrcField   string
	DstField   string
	Conversion string
}

// 生成拷贝函数的模板
const copyFuncTemplate = `// {{.FuncName}} 是一个自动生成的拷贝函数
func {{.FuncName}}({{.DstVar}} *{{.DstType}}, {{.SrcVar}} *{{.SrcType}}) {
{{- range .Fields }}
{{- if .Conversion }}
	{{$.DstVar}}.{{.DstField}} = {{.Conversion}}({{$.SrcVar}}.{{.SrcField}})
{{- else }}
	{{$.DstVar}}.{{.DstField}} = {{$.SrcVar}}.{{.SrcField}}
{{- end }}
{{- end }}
}`

func main() {
	// 要遍历的目录
	dir := "." // 当前目录

	// 遍历目录
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v", path, err)
			return err
		}

		// 只处理 Go 文件
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			log.Printf("Processing file: %s", path)

			// 解析文件
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				log.Printf("Failed to parse file %s: %v", path, err)
				return nil
			}

			// 查找带有 // :quickcopy 注释的函数
			ast.Inspect(file, func(n ast.Node) bool {
				// 查找函数声明
				funcDecl, ok := n.(*ast.FuncDecl)
				if !ok || funcDecl.Doc == nil {
					return true
				}

				// 检查是否有 // :quickcopy 注释
				var isQuickCopy bool
				var allowNarrow bool
				var ignoreCase bool
				var singleToSlice bool
				var fieldMappings map[string]string // 存储字段映射规则
				for _, comment := range funcDecl.Doc.List {
					if strings.Contains(comment.Text, "// :quickcopy") {
						isQuickCopy = true
						// 解析选项
						if strings.Contains(comment.Text, "--allow-narrow") {
							allowNarrow = true
						}
						if strings.Contains(comment.Text, "--ignore-case") {
							ignoreCase = true
						}
						if strings.Contains(comment.Text, "--single-to-slice") {
							singleToSlice = true
						}
						// 解析字段映射规则
						fieldMappings = parseFieldMappings(comment.Text)
						break
					}
				}
				if !isQuickCopy {
					return true
				}

				log.Printf("Found // :quickcopy function: %s", funcDecl.Name.Name)

				// 解析函数签名
				if funcDecl.Type.Params.List == nil || len(funcDecl.Type.Params.List) != 2 {
					log.Fatalf("Copy function %s must have exactly two parameters", funcDecl.Name.Name)
				}

				dstParam := funcDecl.Type.Params.List[0]
				srcParam := funcDecl.Type.Params.List[1]

				srcVar := srcParam.Names[0].Name
				dstVar := dstParam.Names[0].Name

				srcType := strings.TrimPrefix(types.ExprString(srcParam.Type), "*")
				dstType := strings.TrimPrefix(types.ExprString(dstParam.Type), "*")

				log.Printf("Source type: %s, Destination type: %s", srcType, dstType)

				// 提取字段映射关系
				fields := getFieldMappings(srcType, dstType, file, ignoreCase, allowNarrow, singleToSlice, fieldMappings)

				// 生成完整的拷贝函数
				generateCompleteCopyFunc(funcDecl, srcVar, dstVar, srcType, dstType, fields)

				// 将修改后的 AST 写回文件
				writeFile(fset, file, path)
				return true
			})
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Failed to walk directory: %v", err)
	}
}

// parseFieldMappings 解析字段映射规则
func parseFieldMappings(comment string) map[string]string {
	mappings := make(map[string]string)
	// 提取映射规则部分
	start := strings.Index(comment, "// :quickcopy")
	if start == -1 {
		return mappings
	}
	// 去掉注释前缀
	rulePart := strings.TrimSpace(comment[start+len("// :quickcopy"):])
	// 按逗号分割规则
	rules := strings.Split(rulePart, ",")
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}
		// 解析 dstField = srcField
		parts := strings.Split(rule, "=")
		if len(parts) != 2 {
			log.Printf("Invalid mapping rule: %s", rule)
			continue
		}
		dstField := strings.TrimSpace(parts[0])
		srcField := strings.TrimSpace(parts[1])

		// 存储完整的字段路径
		mappings[dstField] = srcField
	}
	return mappings
}

// generateCompleteCopyFunc 生成完整的拷贝函数并替换原始函数
func generateCompleteCopyFunc(funcDecl *ast.FuncDecl, srcVar, dstVar, srcType, dstType string, fields []FieldMapping) {
	// 生成拷贝函数代码
	tmpl, err := template.New("copyFunc").Parse(copyFuncTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	var code bytes.Buffer
	err = tmpl.Execute(&code, CopyFuncInfo{
		FuncName: funcDecl.Name.Name,
		SrcVar:   srcVar,
		DstVar:   dstVar,
		SrcType:  srcType,
		DstType:  dstType,
		Fields:   fields,
	})
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	// 将生成的代码包装在一个完整的 Go 文件中
	wrappedCode := fmt.Sprintf("package main\n\n%s", code.String())

	// 将生成的代码解析为 AST
	fset := token.NewFileSet()
	block, err := parser.ParseFile(fset, "", wrappedCode, parser.ParseComments)
	if err != nil {
		log.Fatalf("Failed to parse generated code: %v", err)
	}

	// 提取生成的函数声明
	newFuncDecl := block.Decls[0].(*ast.FuncDecl)

	// 将原始函数的注释附加到新生成的函数上
	if funcDecl.Doc != nil {
		newFuncDecl.Doc = funcDecl.Doc
	}

	funcDecl.Body = newFuncDecl.Body
}

// writeFile 将修改后的 AST 写回文件
func writeFile(fset *token.FileSet, file *ast.File, path string) {
	// 创建输出文件
	outputFile, err := os.Create(path)
	if err != nil {
		log.Fatalf("Failed to create output file %s: %v", path, err)
	}
	defer outputFile.Close()

	// 格式化整个文件
	var buf bytes.Buffer

	cfg := &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	cfg.Fprint(&buf, fset, file)
	// 将格式化后的内容写入文件
	_, err = outputFile.Write(buf.Bytes())
	if err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	log.Printf("Successfully updated and formatted file: %s", path)
}

// getPackagePathAndTypeName 解析包路径和类型名称
func getPackagePathAndTypeName(typeExpr string) (pkgPath, typeName string) {
	parts := strings.Split(typeExpr, ".")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", typeExpr
}

// findStructDefInPackage 在包中查找结构体定义
func findStructDefInPackage(pkgPath, structName string) (structType *ast.StructType) {
	// 配置 packages.Config
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedDeps,
	}

	// 加载包
	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		log.Printf("Failed to load package %s: %v", pkgPath, err)
		return nil
	}

	if len(pkgs) == 0 {
		log.Printf("No packages found for %s", pkgPath)
		return nil
	}

	// 遍历包中的文件
	for _, pkg := range pkgs {
		log.Printf("Inspecting package: %s, syntax: %d\n", pkg.PkgPath, len(pkg.Syntax))
		for _, file := range pkg.Syntax {
			// 遍历 AST
			ast.Inspect(file, func(n ast.Node) bool {
				// 查找类型声明
				ts, ok := n.(*ast.TypeSpec)
				if !ok {
					return true
				}

				log.Printf("Found type: %s", ts.Name.Name)
				if ts.Name.Name != structName {
					return true // 继续遍历
				}

				// 找到目标结构体
				log.Printf("Found type: %s, struceName:%s\n", ts.Name.Name, structName)
				if structType, ok = ts.Type.(*ast.StructType); ok {
					log.Printf("Type %s is a struct with fields:", ts.Name.Name)
					for _, field := range structType.Fields.List {
						for _, fieldName := range field.Names {
							log.Printf("  - %s", fieldName.Name)
						}
					}
					return false // 退出遍历
				}
				return true
			})
		}
	}

	if structType == nil {
		log.Printf("Struct definition not found: %s in package %s", structName, pkgPath)
	}
	return structType
}

// findStructDef 查找结构体定义
func findStructDef(typeName string, file *ast.File) *ast.StructType {
	// 首先在当前文件中查找
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != typeName {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			log.Printf("Found struct definition: %s", typeName)
			return structType
		}
	}

	// 如果当前文件中没有找到，尝试在包中查找
	pkgPath, typeName := getPackagePathAndTypeName(typeName)
	if pkgPath == "" {
		log.Printf("Struct definition not found: %s", typeName)
		return nil
	}

	// 从文件的 imports 中查找完整的包路径
	var fullPkgPath string
	for _, imp := range file.Imports {
		impPath := strings.Trim(imp.Path.Value, `"`)
		if strings.HasSuffix(impPath, pkgPath) {
			fullPkgPath = impPath
			break
		}
	}

	if fullPkgPath == "" {
		log.Printf("Package path not found in imports: %s", pkgPath)
		return nil
	}

	structType := findStructDefInPackage(fullPkgPath, typeName)
	if structType == nil {
		log.Printf("Struct definition not found: %s in package %s", typeName, fullPkgPath)
		return nil
	}

	return structType
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

// getFieldMappings 获取字段映射关系
func getFieldMappings(srcType, dstType string, file *ast.File, ignoreCase, allowNarrow, singleToSlice bool, fieldMappings map[string]string) []FieldMapping {
	var fields []FieldMapping

	// 查找源类型和目标类型的结构体定义
	srcStruct := findStructDef(srcType, file)
	dstStruct := findStructDef(dstType, file)

	if srcStruct == nil || dstStruct == nil {
		log.Fatalf("Failed to find struct definitions for %s or %s", srcType, dstType)
	}

	log.Printf("Found struct definitions: %s and %s", srcType, dstType)

	// 用于记录已经映射的目标字段
	mappedDstFields := make(map[string]bool)

	// 如果有显式的字段映射规则，则按照规则进行映射
	for dstFieldPath, srcFieldPath := range fieldMappings {
		// 查找目标字段
		dstFieldName := extractFieldName(dstFieldPath)
		dstField := findFieldByName(dstStruct, dstFieldName)
		if dstField == nil {
			log.Printf("Destination field not found: %s", dstFieldName)
			continue
		}

		// 查找源字段
		srcFieldName := extractFieldName(srcFieldPath)
		srcField := findFieldByName(srcStruct, srcFieldName)
		if srcField == nil {
			log.Printf("Source field not found: %s", srcFieldName)
			continue
		}

		// 获取类型转换逻辑
		conversion := getTypeConversion(types.ExprString(srcField.Type), types.ExprString(dstField.Type), allowNarrow, singleToSlice)

		// 存储映射关系
		fields = append(fields, FieldMapping{
			SrcField:   srcFieldPath, // 使用完整的源字段路径
			DstField:   dstFieldPath, // 使用完整的目标字段路径
			Conversion: conversion,
		})
		log.Printf("Mapped field: %s -> %s (Conversion: %s)", srcFieldPath, dstFieldPath, conversion)

		// 标记该目标字段已经映射
		mappedDstFields[dstFieldName] = true
	}

	// 遍历目标结构体的字段，处理未映射的字段
	for _, dstField := range dstStruct.Fields.List {
		for _, dstFieldName := range dstField.Names {
			// 如果目标字段已经映射过，则跳过
			if mappedDstFields[dstFieldName.Name] {
				continue
			}

			// 查找源结构体中是否有同名字段（支持忽略大小写）
			var srcField *ast.Field
			var srcFieldName string
			if ignoreCase {
				srcField, srcFieldName = findFieldByNameIgnoreCase(srcStruct, dstFieldName.Name)
			} else {
				srcField = findFieldByName(srcStruct, dstFieldName.Name)
				if srcField != nil {
					srcFieldName = srcField.Names[0].Name
				}
			}

			if srcField == nil {
				log.Printf("Source field not found for destination field: %s", dstFieldName.Name)
				continue
			}

			// 获取类型转换逻辑
			conversion := getTypeConversion(types.ExprString(srcField.Type), types.ExprString(dstField.Type), allowNarrow, singleToSlice)

			// 存储映射关系
			fields = append(fields, FieldMapping{
				SrcField:   srcFieldName,      // 使用源字段的原始名称
				DstField:   dstFieldName.Name, // 使用目标字段的原始名称
				Conversion: conversion,
			})
			log.Printf("Mapped field: %s -> %s (Conversion: %s)", srcFieldName, dstFieldName.Name, conversion)
		}
	}

	return fields
}

// findFieldByNameIgnoreCase 在结构体中查找字段（忽略大小写），并返回匹配的字段及其原始名称
func findFieldByNameIgnoreCase(structType *ast.StructType, fieldName string) (*ast.Field, string) {
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			if strings.EqualFold(name.Name, fieldName) {
				return field, name.Name // 返回字段及其原始名称
			}
		}
	}
	return nil, ""
}

// findFieldByName 在结构体中查找字段
func findFieldByName(structType *ast.StructType, fieldName string) *ast.Field {
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			if name.Name == fieldName {
				return field
			}
		}
	}
	return nil
}

// getTypeConversion 获取类型转换逻辑
func getTypeConversion(srcType, dstType string, allowNarrow, singleToSlice bool) string {
	// 处理数组和切片类型（仅在启用 singleToSlice 时）
	if singleToSlice && (strings.HasPrefix(dstType, "[]") || strings.Contains(dstType, "[")) {
		// 如果目标类型是切片或数组，且源类型是单个元素类型
		if !strings.HasPrefix(srcType, "[]") && !strings.Contains(srcType, "[") {
			// 获取目标类型的元素类型
			dstElemType := strings.TrimPrefix(dstType, "[]")
			dstElemType = strings.Split(dstElemType, "[")[0]

			// 如果源类型和目标元素类型相同，则直接赋值
			if srcType == dstElemType {
				return fmt.Sprintf("func(src %s) %s { return []%s{src} }", srcType, dstType, dstElemType)
			}

			// 如果需要进行类型转换，则生成相应的转换函数
			conversion := getTypeConversion(srcType, dstElemType, allowNarrow, singleToSlice)
			if conversion != "" {
				return fmt.Sprintf("func(src %s) %s { return []%s{%s(src)} }", srcType, dstType, dstElemType, conversion)
			}
		}
	}

	// 处理整数类型转换
	if isIntegerType(srcType) && isIntegerType(dstType) {
		srcWidth := getIntWidth(srcType)
		dstWidth := getIntWidth(dstType)

		// 如果源类型和目标类型不同，则需要显式转换
		if srcType != dstType {
			// 如果是窄化转换（高宽度到低宽度），检查是否允许
			if srcWidth > dstWidth {
				if !allowNarrow {
					log.Printf("Warning: Narrowing conversion from %s to %s is not allowed", srcType, dstType)
					return "" // 不允许窄化转换
				}
				log.Printf("Warning: Narrowing conversion from %s to %s may lose data", srcType, dstType)
			}
			return dstType
		}
	}

	// 处理其他类型的转换
	switch {
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
	default:
		return ""
	}
}

func isIntegerType(typeName string) bool {
	switch typeName {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
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
	default:
		return 0
	}
}
