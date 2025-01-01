package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// CopyFuncInfo 存储拷贝函数信息
type CopyFuncInfo struct {
	FuncName string
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
const copyFuncTemplate = `{{- range .Fields }}
{{- if .Conversion }}
	dst.{{.DstField}} = {{.Conversion}}(src.{{.SrcField}})
{{- else }}
	dst.{{.DstField}} = src.{{.SrcField}}
{{- end }}
{{- end }}`

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

				srcParam := funcDecl.Type.Params.List[0]
				dstParam := funcDecl.Type.Params.List[1]

				srcType := strings.TrimPrefix(types.ExprString(srcParam.Type), "*")
				dstType := strings.TrimPrefix(types.ExprString(dstParam.Type), "*")

				log.Printf("Source type: %s, Destination type: %s", srcType, dstType)

				// 提取字段映射关系
				fields := getFieldMappings(srcType, dstType, file)

				// 生成拷贝函数实现
				generateCopyFuncImpl(funcDecl, fields)

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

// generateCopyFuncImpl 生成拷贝函数实现并插入到函数体中
func generateCopyFuncImpl(funcDecl *ast.FuncDecl, fields []FieldMapping) {
	// 生成拷贝函数代码
	tmpl, err := template.New("copyFunc").Parse(copyFuncTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	var code bytes.Buffer
	err = tmpl.Execute(&code, CopyFuncInfo{
		FuncName: funcDecl.Name.Name,
		SrcType:  types.ExprString(funcDecl.Type.Params.List[0].Type),
		DstType:  types.ExprString(funcDecl.Type.Params.List[1].Type),
		Fields:   fields,
	})
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	// 将生成的代码包装在一个完整的 Go 文件中
	wrappedCode := fmt.Sprintf("package main\n\nfunc temp() {\n%s\n}", code.String())

	// 将生成的代码解析为 AST
	fset := token.NewFileSet()
	block, err := parser.ParseFile(fset, "", wrappedCode, parser.ParseComments)
	if err != nil {
		log.Fatalf("Failed to parse generated code: %v", err)
	}

	// 提取函数体中的语句
	tempFunc := block.Decls[0].(*ast.FuncDecl)
	funcDecl.Body.List = tempFunc.Body.List
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
	err = format.Node(&buf, fset, file)
	if err != nil {
		log.Fatalf("Failed to format file: %v", err)
	}

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

// findStructDef 查找结构体定义
func findStructDef(typeName string, file *ast.File) *ast.StructType {
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

	log.Printf("Struct definition not found: %s", typeName)
	return nil
}

// getTypeConversion 获取类型转换逻辑
func getTypeConversion(srcType, dstType string) string {
	switch {
	case srcType == "int" && dstType == "string":
		return "fmt.Sprint"
	case srcType == "string" && dstType == "int":
		return "func(s string) int { i, _ := strconv.Atoi(s); return i }"
	default:
		return ""
	}
}
