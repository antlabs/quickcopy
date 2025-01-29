package quickcopy

import (
	"go/ast"
	"go/token"
	"log"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// 新增缓存：包路径 -> 包语法树
// var pkgCache = make(map[string][]*ast.File)
var pkgCache Map[string, []*ast.File]

// findStructDef 优化后的实现
func findStructDef(typeName string, file *ast.File, path string) *ast.StructType {
	if file == nil {
		log.Printf("File is nil, cannot find struct definition for %s", typeName)
		return nil
	}

	// 1. 尝试从当前文件查找
	if structType := findInCurrentFile(typeName, file); structType != nil {
		return structType
	}

	// 2. 解析包路径和类型名（处理形如 pkg.Type 的类型）
	pkgPath, typeName := parsePkgType(typeName)
	if pkgPath != "" {
		// 处理外部包类型
		return findInImportedPackage(pkgPath, typeName, file)
	}

	// 3. 尝试从同一包的其他文件查找
	return findInCurrentPackage(typeName, file, path)
}

// 1. 在当前文件查找
func findInCurrentFile(typeName string, file *ast.File) *ast.StructType {
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

			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				log.Printf("Found struct in current file: %s", typeName)
				return structType
			}
		}
	}
	return nil
}

// 2. 在导入的包中查找
func findInImportedPackage(pkgPath, typeName string, file *ast.File) *ast.StructType {
	// 从文件的 imports 中查找完整的包路径
	var fullPkgPath string
	for _, imp := range file.Imports {
		importedPath := strings.Trim(imp.Path.Value, `"`)
		// 匹配别名或路径最后部分
		if imp.Name != nil && imp.Name.Name == pkgPath {
			fullPkgPath = importedPath
			break
		}
		if filepath.Base(importedPath) == pkgPath {
			fullPkgPath = importedPath
			break
		}
	}

	if fullPkgPath == "" {
		log.Printf("Package not imported: %s", pkgPath)
		return nil
	}

	return findStructDefInPackage(fullPkgPath, typeName)
}

// 3. 在同一包的其他文件中查找
func findInCurrentPackage(typeName string, file *ast.File, path string) *ast.StructType {

	pkgPath := filepath.Dir(path)

	// 检查缓存

	files, ok := pkgCache.Load(pkgPath)

	if !ok {
		// 未缓存，加载包信息
		cfg := &packages.Config{
			Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax,
			Dir:  pkgPath,
		}
		pkgs, err := packages.Load(cfg, ".")
		if err != nil || len(pkgs) == 0 {
			log.Printf("Failed to load current package: %v", err)
			return nil
		}

		// 提取语法树并缓存
		var syntax []*ast.File
		for _, pkg := range pkgs {
			syntax = append(syntax, pkg.Syntax...)
		}

		pkgCache.Store(pkgPath, syntax)
		files = syntax
	}

	// 遍历所有文件查找结构体
	for _, f := range files {
		if structType := findInCurrentFile(typeName, f); structType != nil {
			log.Printf("Found struct in package %s: %s", pkgPath, typeName)
			return structType
		}
	}

	return nil
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

func parsePkgType(typeStr string) (pkgPath, typeName string) {
	parts := strings.Split(typeStr, ".")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], "."), parts[len(parts)-1]
	}
	return "", typeStr
}
