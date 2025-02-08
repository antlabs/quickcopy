package quickcopy

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strings"
)

func addRequiredImports(file *ast.File, importPath ...string) {

	log.Printf("addRequiredImports:%v\n", importPath)

	if len(importPath) == 0 {
		return
	}
	// 需要添加的包
	requiredImports := make(map[string]bool)

	for _, pkg := range importPath {
		requiredImports[pkg] = true
	}

	// 查找现有的 import 声明
	var importDecl *ast.GenDecl
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			importDecl = genDecl
			break
		}
	}

	// 如果没有 import 声明，创建一个新的
	if importDecl == nil {
		importSpecs := make([]ast.Spec, 0, len(requiredImports))
		for pkg := range requiredImports {
			importSpecs = append(importSpecs, &ast.ImportSpec{
				Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, pkg)},
			})
		}

		// 将新的 import 声明添加到文件顶部
		file.Decls = append([]ast.Decl{&ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: importSpecs,
		}}, file.Decls...)
		return
	}

	// 如果已有 import 声明，检查并添加缺失的包
	existingImports := make(map[string]bool)
	for _, spec := range importDecl.Specs {
		if importSpec, ok := spec.(*ast.ImportSpec); ok {
			existingImports[strings.Trim(importSpec.Path.Value, `"`)] = true
		}
	}

	// 添加缺失的包
	for pkg := range requiredImports {
		if !existingImports[pkg] {
			log.Printf("Adding import: %s", pkg)
			importDecl.Specs = append(importDecl.Specs, &ast.ImportSpec{
				Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, pkg)},
			})
		}
	}
}
