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
				for _, comment := range funcDecl.Doc.List {
					if strings.Contains(comment.Text, "// :quickcopy") {
						isQuickCopy = true
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
				fields := getFieldMappings(srcType, dstType, file)

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

// getFieldMappings 获取字段映射关系
func getFieldMappings(srcType, dstType string, file *ast.File) []FieldMapping {
	var fields []FieldMapping

	// 查找源类型和目标类型的结构体定义
	srcStruct := findStructDef(srcType, file)
	dstStruct := findStructDef(dstType, file)

	if srcStruct == nil || dstStruct == nil {
		log.Fatalf("Failed to find struct definitions for %s or %s", srcType, dstType)
	}

	log.Printf("Found struct definitions: %s and %s", srcType, dstType)

	// 提取字段映射关系
	for _, srcField := range srcStruct.Fields.List {
		for _, dstField := range dstStruct.Fields.List {
			if srcField.Names[0].Name == dstField.Names[0].Name {
				conversion := getTypeConversion(types.ExprString(srcField.Type), types.ExprString(dstField.Type))
				fields = append(fields, FieldMapping{
					SrcField:   srcField.Names[0].Name,
					DstField:   dstField.Names[0].Name,
					Conversion: conversion,
				})
				log.Printf("Mapped field: %s -> %s (Conversion: %s)", srcField.Names[0].Name, dstField.Names[0].Name, conversion)
			}
		}
	}

	return fields
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

// getTypeConversion 获取类型转换逻辑
func getTypeConversion(srcType, dstType string) string {
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
