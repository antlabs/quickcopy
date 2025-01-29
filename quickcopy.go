package quickcopy

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

// 新增函数：注册生成的函数到AST
var generatedFunctions Map[string, *ast.FuncDecl]

var processedTopLevelTypes Map[string, bool]

// FieldMapping 增加新字段
type FieldMapping struct {
	SrcField       string
	DstField       string
	Conversion     string
	IsSlice        bool
	IsEmbedded     bool
	ConversionFunc string
	SrcElemType    string // 新增
	DstElemType    string // 新增
}

// CopyFuncInfo 存储拷贝函数信息
type CopyFuncInfo struct {
	FuncName string
	SrcVar   string
	DstVar   string
	SrcType  string
	DstType  string
	Fields   []FieldMapping
}

// 修改后的完整 processFields 函数
func processFields(
	structType *ast.StructType,
	srcStruct *ast.StructType,
	file *ast.File,
	prefix string,
	isSrc bool,
	ignoreCase bool,
	allowNarrow bool,
	singleToSlice bool,
	fields *[]FieldMapping,
	mappedDstFields map[string]bool,
) {
	for _, field := range structType.Fields.List {
		// 处理嵌入字段（匿名和具名）
		if len(field.Names) == 0 {
			// 匿名嵌入字段
			embeddedType := types.ExprString(field.Type)
			if embeddedStruct := findStructDef(embeddedType, file); embeddedStruct != nil {
				processFields(
					embeddedStruct,
					srcStruct,
					file,
					prefix, // 保持当前前缀实现字段提升
					isSrc,
					ignoreCase,
					allowNarrow,
					singleToSlice,
					fields,
					mappedDstFields,
				)
			}
			continue
		}

		// 处理具名字段（包括具名嵌入）
		for _, fieldName := range field.Names {
			currentFieldPath := prefix + fieldName.Name
			fieldType := types.ExprString(field.Type)

			// 检查是否是具名嵌入结构体
			if isStructType(fieldType, file) {
				if embeddedStruct := findStructDef(fieldType, file); embeddedStruct != nil {
					processFields(
						embeddedStruct,
						srcStruct,
						file,
						currentFieldPath+".", // 添加点号保持嵌套路径
						isSrc,
						ignoreCase,
						allowNarrow,
						singleToSlice,
						fields,
						mappedDstFields,
					)
				}
			}

			// 普通字段处理逻辑
			var srcField *ast.Field
			var srcFieldName string
			if isSrc {
				srcField = field
				srcFieldName = currentFieldPath
			} else {
				if ignoreCase {
					srcField, srcFieldName = findFieldByNameIgnoreCase(srcStruct, fieldName.Name, file)
				} else {
					srcField = findFieldByName(srcStruct, fieldName.Name, file)
					if srcField != nil {
						srcFieldName = srcField.Names[0].Name
					}
				}
			}

			if srcField == nil {
				continue
			}

			// 处理类型转换
			srcType := types.ExprString(srcField.Type)
			dstType := types.ExprString(field.Type)
			conversion := getTypeConversion(srcType, dstType, allowNarrow, singleToSlice, file)

			// 判断是否为嵌入字段
			isEmbedded := false
			if isStructType(srcType, file) && isStructType(dstType, file) {
				if srcType == dstType {
					isEmbedded = true
					conversion = getStructCopyFuncName(srcType, dstType)
				}
			}

			// 存储映射关系
			*fields = append(*fields, FieldMapping{
				SrcField:       srcFieldName,
				DstField:       currentFieldPath,
				Conversion:     conversion,
				IsEmbedded:     isEmbedded,
				ConversionFunc: getStructCopyFuncName(srcType, dstType),
				SrcElemType:    getElementType(srcType),
				DstElemType:    getElementType(dstType),
			})

			if !isSrc {
				mappedDstFields[fieldName.Name] = true
			}
		}
	}
}

// 新增 processNestedTypes 函数
func processNestedTypes(structType *ast.StructType, file *ast.File) {
	for _, field := range structType.Fields.List {
		fieldType := types.ExprString(field.Type)

		// 处理指针类型
		if strings.HasPrefix(fieldType, "*") {
			elemType := strings.TrimPrefix(fieldType, "*")
			if isStructType(elemType, file) {
				generateCopyFunctionIfNeeded(elemType, elemType, file)
			}
			continue
		}

		// 处理切片/数组类型
		if isSliceOrArray(fieldType) {
			elemType := getElementType(fieldType)
			generateCopyFunctionIfNeeded(elemType, elemType, file)
			continue
		}

		// 处理嵌套结构体
		if isStructType(fieldType, file) {
			generateCopyFunctionIfNeeded(fieldType, fieldType, file)
		}
	}
}

// 新增 isExternalType 函数
func isExternalType(typeName string, file *ast.File) bool {
	// 处理指针类型
	if strings.HasPrefix(typeName, "*") {
		return isExternalType(strings.TrimPrefix(typeName, "*"), file)
	}

	// 处理包前缀
	if strings.Contains(typeName, ".") {
		pkgPart := strings.Split(typeName, ".")[0]
		for _, imp := range file.Imports {
			importedPath := strings.Trim(imp.Path.Value, `"`)
			// 匹配导入路径最后部分或别名
			if imp.Name != nil && imp.Name.Name == pkgPart {
				return true
			}
			if filepath.Base(importedPath) == pkgPart {
				return true
			}
		}
	}
	return false
}

var generatedStructPairs Map[string, bool]

func generateCopyFunctionIfNeeded(srcType, dstType string, file *ast.File) {
	// 增强过滤条件：只有当类型不同时才生成函数
	// if srcType == dstType {
	// 	return
	// }

	key := srcType + "->" + dstType
	if _, ok := processedTopLevelTypes.Load(key); ok {
		return
	}
	// 防止死循环的。
	if _, loaded := generatedStructPairs.LoadOrStore(key, true); loaded {
		return
	}
	funcName := getStructCopyFuncName(srcType, dstType)
	if _, ok := generatedFunctions.Load(funcName); ok {
		return
	}

	// 新增过滤逻辑：跳过基本类型和外部包类型
	if isBasicType(srcType) || isBasicType(dstType) {
		return
	}

	if isExternalType(srcType, file) || isExternalType(dstType, file) {
		return
	}

	// // 新增：跳过非结构体类型
	// if !isStructType(srcType, file) || !isStructType(dstType, file) {
	// 	return
	// }

	srcStruct := findStructDef(srcType, file)
	dstStruct := findStructDef(dstType, file)
	if srcStruct == nil || dstStruct == nil {
		return
	}

	// 递归处理所有嵌套类型
	processNestedTypes(srcStruct, file)
	processNestedTypes(dstStruct, file)

	fields := getFieldMappings(srcType, dstType, file, false, false, false, nil)

	funcDecl := &ast.FuncDecl{
		Name: ast.NewIdent(funcName),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{Names: []*ast.Ident{ast.NewIdent("dst")}, Type: &ast.StarExpr{X: ast.NewIdent(dstType)}},
					{Names: []*ast.Ident{ast.NewIdent("src")}, Type: &ast.StarExpr{X: ast.NewIdent(srcType)}},
				},
			},
		},
	}
	generateCompleteCopyFunc(funcDecl, "src", "dst", srcType, dstType, fields)
	// 注册生成的函数
	generatedFunctions.Store(funcName, funcDecl)
}

const copyFuncTemplate = `// {{.FuncName}} 是一个自动生成的拷贝函数
func {{.FuncName}}({{.DstVar}} *{{.DstType}}, {{.SrcVar}} *{{.SrcType}}) {
{{- range .Fields }}
{{- if .IsSlice }}
    // 处理切片字段 {{.DstField}}
    {{$.DstVar}}.{{.DstField}} = {{.Conversion}}({{$.SrcVar}}.{{.SrcField}})
{{- else if .IsEmbedded }}
    // 处理嵌入字段 {{.DstField}}
    {{if .ConversionFunc -}}
    {{.ConversionFunc}}({{$.DstVar}}.{{.DstField}}, {{$.SrcVar}}.{{.SrcField}})
    {{- else -}}
    {{$.DstVar}}.{{.DstField}} = {{$.SrcVar}}.{{.SrcField}} // 默认直接拷贝
    {{- end}}
{{- else if .Conversion }}
    // 类型转换字段 {{.DstField}}
    {{$.DstVar}}.{{.DstField}} = {{.Conversion}}({{$.SrcVar}}.{{.SrcField}})
{{- else }}
    // 直接赋值字段 {{.DstField}}
    {{$.DstVar}}.{{.DstField}} = {{$.SrcVar}}.{{.SrcField}}
{{- end }}
{{- end }}
}`

func addGeneratedFunction(funcName string, fn *ast.FuncDecl) {
	log.Printf("Adding generated function: %s", funcName)
	generatedFunctions.Store(funcName, fn)
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
		log.Fatalf("Failed to parse generated code: %v, %s", err, wrappedCode)
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
	var existingFuncs Map[string, int]
	// 构建现有函数索引（名称 -> 声明位置）
	for i, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			fmt.Printf("Found existing function: %s\n", fn.Name.Name)
			existingFuncs.Store(fn.Name.Name, i)
		}
	}

	// 合并生成的函数
	generatedFunctions.Range(func(_ string, newFn *ast.FuncDecl) bool {
		name := newFn.Name.Name
		log.Printf("Processing function: %s", name)
		if idx, exists := existingFuncs.Load(name); exists {
			// 替换已存在的函数声明
			file.Decls[idx] = newFn
		} else {
			// 追加新的函数声明
			file.Decls = append(file.Decls, newFn)
		}
		return true
	})

	// 清空注册表
	generatedFunctions.Clear()

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

// findStructDef 查找结构体定义，支持内嵌结构体
func findStructDef(typeName string, file *ast.File) *ast.StructType {
	// 如果 file 为 nil，直接返回（避免空指针）
	if file == nil {
		log.Printf("File is nil, cannot find struct definition for %s", typeName)
		return nil
	}
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

// getFieldMappings 获取字段映射关系，支持结构体内嵌
func getFieldMappings(srcType, dstType string, file *ast.File, ignoreCase, allowNarrow, singleToSlice bool, fieldMappings map[string]string) []FieldMapping {
	// 生成内层结构体的拷贝函数（如果存在）
	generateCopyFunctionIfNeeded(srcType, dstType, file)
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
		dstField := findFieldByName(dstStruct, dstFieldName, file)
		if dstField == nil {
			log.Printf("Destination field not found: %s", dstFieldName)
			continue
		}

		// 查找源字段
		srcFieldName := extractFieldName(srcFieldPath)
		srcField := findFieldByName(srcStruct, srcFieldName, file)
		if srcField == nil {
			log.Printf("Source field not found: %s", srcFieldName)
			continue
		}

		// 获取类型转换逻辑
		conversion := getTypeConversion(types.ExprString(srcField.Type), types.ExprString(dstField.Type), allowNarrow, singleToSlice, file)

		// 判断是否为嵌入字段
		isEmbedded := isEmbeddedField(srcField) || isEmbeddedField(dstField)
		// 存储映射关系
		fields = append(fields, FieldMapping{
			SrcField:       srcFieldPath, // 使用完整的源字段路径
			DstField:       dstFieldPath, // 使用完整的目标字段路径
			Conversion:     conversion,
			IsEmbedded:     isEmbedded,
			ConversionFunc: getStructCopyFuncName(srcType, dstType),
		})
		log.Printf("Mapped field: %s -> %s (Conversion: %s)", srcFieldPath, dstFieldPath, conversion)

		// 标记该目标字段已经映射
		mappedDstFields[dstFieldName] = true
	}

	// 处理目标结构体的字段
	processFields(dstStruct, srcStruct, file, "", false, ignoreCase, allowNarrow, singleToSlice, &fields, mappedDstFields)
	return fields
}

// 新增辅助函数：判断字段是否为嵌入字段
func isEmbeddedField(field *ast.Field) bool {
	return len(field.Names) == 0
}

// findFieldByName 在结构体中查找字段，支持内嵌结构体
func findFieldByName(structType *ast.StructType, fieldName string, file *ast.File) *ast.Field {
	for _, field := range structType.Fields.List {
		// 处理内嵌结构体
		if len(field.Names) == 0 {
			// 内嵌结构体
			embeddedType := types.ExprString(field.Type)
			// 关键修改：传递当前文件的 AST 节点
			embeddedStruct := findStructDef(embeddedType, file)
			if embeddedStruct != nil {
				// 递归查找内嵌结构体的字段
				if foundField := findFieldByName(embeddedStruct, fieldName, file); foundField != nil {
					return foundField
				}
			}
			continue
		}

		// 处理普通字段
		for _, name := range field.Names {
			if name.Name == fieldName {
				return field
			}
		}
	}
	return nil
}

// findFieldByNameIgnoreCase 在结构体中查找字段（忽略大小写），支持内嵌结构体
func findFieldByNameIgnoreCase(structType *ast.StructType, fieldName string, file *ast.File) (*ast.Field, string) {
	for _, field := range structType.Fields.List {
		// 处理内嵌结构体
		if len(field.Names) == 0 {
			// 内嵌结构体
			embeddedType := types.ExprString(field.Type)
			// 关键修改：传递当前文件的 AST 节点
			embeddedStruct := findStructDef(embeddedType, file)
			if embeddedStruct != nil {
				// 递归查找内嵌结构体的字段
				if foundField, foundName := findFieldByNameIgnoreCase(embeddedStruct, fieldName, file); foundField != nil {
					return foundField, foundName
				}
			}
			continue
		}

		// 处理普通字段
		for _, name := range field.Names {
			if strings.EqualFold(name.Name, fieldName) {
				return field, name.Name
			}
		}
	}
	return nil, ""
}

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

// 新增函数：获取切片拷贝函数名
func getSliceCopyFuncName(srcElem, dstElem string) string {
	return fmt.Sprintf("copySlice%sFromSlice%s",
		sanitizeTypeName(dstElem),
		sanitizeTypeName(srcElem))
}

// 修改 getTypeConversion 函数签名，增加 file 参数
func getTypeConversion(srcType, dstType string, allowNarrow, singleToSlice bool, file *ast.File) string {
	// 类型相同无需转换
	if srcType == dstType {
		return ""
	}

	// 处理切片转换（新增部分）
	if isSliceOrArray(srcType) && isSliceOrArray(dstType) {
		return handleSliceConversion(srcType, dstType, allowNarrow, singleToSlice, file)
	}

	// 处理基本类型转换
	if isBasicType(srcType) && isBasicType(dstType) {
		return handleBasicConversion(srcType, dstType, allowNarrow)
	}
	// 处理结构体类型
	if isStructType(srcType, file) && isStructType(dstType, file) {
		return generateStructConversionFunc(srcType, dstType)
	}

	// 处理指针类型
	if isPointerType(srcType) && isPointerType(dstType) {
		return handlePointerConversion(srcType, dstType, allowNarrow, singleToSlice, file)
	}

	// 其他类型转换逻辑
	return handleSpecialTypeConversion(srcType, dstType)
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

func handleSliceConversion(srcType, dstType string, allowNarrow, singleToSlice bool, file *ast.File) string {

	srcElem := getElementType(srcType)
	dstElem := getElementType(dstType)

	// 生成元素转换函数
	elemConv := getStructCopyFuncName(srcElem, dstElem)
	generateCopyFunctionIfNeeded(srcElem, dstElem, file)

	// 只有当元素类型需要转换时才生成切片函数
	if elemConv != "" {
		log.Printf("Generating slice conversion function for %s to %s, funcName: %s", srcType, dstType, elemConv)
		return generateSliceCopyFunc(srcElem, dstElem, elemConv, file)
	}
	return "" // 直接赋值

}

// 修改 generateSliceCopyFunc 函数
func generateSliceCopyFunc(srcElem, dstElem, elemConv string, file *ast.File) string {
	// 强制生成元素类型的转换函数
	generateCopyFunctionIfNeeded(srcElem, dstElem, file)

	funcName := getSliceCopyFuncName(srcElem, dstElem)

	// 如果元素类型相同，直接返回浅拷贝
	if srcElem == dstElem {
		return fmt.Sprintf("func(src []%s) []%s { return append([]%s(nil), src...) }", srcElem, dstElem, dstElem)
	}
	// 使用 LoadOrStore 确保并发安全
	if _, loaded := generatedFunctions.Load(funcName); loaded {
		log.Printf("Slice function %s already generated", funcName)
		return funcName
	}

	// 处理file为nil的情况
	srcIsStruct := file != nil && isStructType(srcElem, file)
	dstIsStruct := file != nil && isStructType(dstElem, file)

	// 如果源和目标元素都不是结构体，生成直接拷贝函数
	if !srcIsStruct && !dstIsStruct {
		return fmt.Sprintf("func(src []%s) []%s { return src }", srcElem, dstElem)
	}

	// 需要转换函数时，确保elemConv非空
	if elemConv == "" {
		log.Printf("elemConv is required for struct elements but is empty")
		return ""
	}

	code := fmt.Sprintf(`
	package main
// %s 是自动生成的切片拷贝函数
func %s(src []%s) []%s {
	if src == nil {
		return nil
	}
	dst := make([]%s, len(src))
	for i := range src {
		%s(&dst[i], &src[i])
	}
	return dst
}`, funcName, funcName, srcElem, dstElem, dstElem, elemConv)

	// 安全解析生成的代码
	fset := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		log.Printf("Failed to parse generated slice function: %v", err)
		return ""
	}

	if len(parsedFile.Decls) == 0 {
		log.Printf("Generated slice function is empty")
		return ""
	}

	if fn, ok := parsedFile.Decls[0].(*ast.FuncDecl); ok {
		addGeneratedFunction(funcName, fn)
		return funcName
	}
	return ""
}

// isStructType 判断给定类型是否为结构体类型（包含指针类型和跨包类型）
func isStructType(typeName string, file *ast.File) bool {
	if typeName == "time.Time" {
		return false
	}
	// 快速跳过基本类型和复合类型
	if isBasicType(typeName) || isSliceOrArray(typeName) || isPointerType(typeName) {
		return false
	}

	// 处理指针类型（递归判断底层类型）
	if strings.HasPrefix(typeName, "*") {
		return isStructType(strings.TrimPrefix(typeName, "*"), file)
	}

	// 处理数组/切片前缀（递归判断元素类型）

	if strings.HasPrefix(typeName, "[]") || strings.Contains(typeName, "[") {
		elemType := getElementType(typeName)
		return isStructType(elemType, file)
	}

	// 分解包前缀（处理形如 pkg.Struct 的类型）
	pkgPath, typeName := parsePkgType(typeName)

	// 在当前文件的 imports 中查找匹配的包路径
	if pkgPath != "" {
		for _, imp := range file.Imports {
			importedPath := strings.Trim(imp.Path.Value, `"`)
			if importedPath == pkgPath ||
				(imp.Name != nil && imp.Name.Name == pkgPath) {
				return findStructDefInPackage(importedPath, typeName) != nil
			}
		}
	}

	// 查找本地结构体定义
	return findStructDef(typeName, file) != nil
}

// parsePkgType 分解包路径和类型名（支持 . 分隔符）
// 示例：
//
//	"pkg.User" → ("pkg", "User")
//	"User" → ("", "User")
func parsePkgType(typeStr string) (pkgPath, typeName string) {
	parts := strings.Split(typeStr, ".")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], "."), parts[len(parts)-1]
	}
	return "", typeStr
}

// 核心处理函数
func handleBasicConversion(src, dst string, allowNarrow bool) string {
	// 整数类型转换
	if isIntegerType(src) && isIntegerType(dst) {
		srcWidth := getIntWidth(src)
		dstWidth := getIntWidth(dst)

		if srcWidth > dstWidth && !allowNarrow {
			log.Printf("Narrowing conversion disabled: %s -> %s", src, dst)
			return ""
		}
		return dst // 返回类型名称作为转换函数
	}

	// 其他基本类型转换
	return handleSpecialTypeConversion(src, dst)
}

func generateElementConversion(srcVar, dstVar, conversion string) string {
	if conversion == "" {
		return fmt.Sprintf("%s = %s", dstVar, srcVar)
	}

	// 处理结构体转换
	if strings.HasPrefix(conversion, "func(") {
		funcName := getStructCopyFuncName(
			strings.TrimPrefix(getElementType(srcVar), "*"),
			strings.TrimPrefix(getElementType(dstVar), "*"),
		)
		return fmt.Sprintf("%s(&%s, &%s)", funcName, dstVar, srcVar)
	}

	// 基本类型转换
	return fmt.Sprintf("%s = %s(%s)", dstVar, conversion, srcVar)
}

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

func getStructCopyFuncName(src, dst string) string {
	// if src == dst {
	// 	return "" // 相同类型不需要转换函数
	// }
	srcClean := sanitizeTypeName(src)
	dstClean := sanitizeTypeName(dst)
	return fmt.Sprintf("copy%sFrom%s", dstClean, srcClean)
}

func isSliceOrArray(t string) bool {
	return strings.Contains(t, "[") || strings.HasPrefix(t, "[]")
}

// getElementType 获取容器类型的元素类型
func getElementType(containerType string) string {
	// 处理数组类型（如 [3]int → int）
	if strings.Contains(containerType, "]") {
		return containerType[strings.Index(containerType, "]")+1:]
	}

	// 处理切片类型（如 []int → int）
	if strings.HasPrefix(containerType, "[]") {
		return strings.TrimPrefix(containerType, "[]")
	}

	return containerType
}

func generateStructConversionFunc(src, dst string) string {
	funcName := getStructCopyFuncName(src, dst)
	return fmt.Sprintf(`%s(dst *%s, src *%s) {
        // Auto-generated struct copy
        // Implement field mappings here
    }`, funcName, dst, src)
}

// 新增指针类型判断函数
func isPointerType(typeName string) bool {
	return strings.HasPrefix(typeName, "*")
}

// 新增指针转换处理函数
func handlePointerConversion(srcType, dstType string, allowNarrow, singleToSlice bool, file *ast.File) string {
	// 获取基础类型
	baseSrc := strings.TrimPrefix(srcType, "*")
	baseDst := strings.TrimPrefix(dstType, "*")

	// 递归获取基础类型转换
	baseConv := getTypeConversion(baseSrc, baseDst, allowNarrow, singleToSlice, file)

	// 生成指针转换逻辑
	return fmt.Sprintf(`func(src %s) %s {
        if src == nil {
            return nil
        }
        dst := new(%s)
        %s
        return dst
    }`, srcType, dstType, baseDst, generateElementConversion("*src", "dst", baseConv))
}

func Main(dir string) {
	// 要遍历的目录
	if dir == "" {
		dir = "." // 当前目录
	}

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

				processedTopLevelTypes.Store(fmt.Sprintf("%s->%s", srcType, dstType), true)

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
