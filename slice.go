package quickcopy

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

func generateBasicSliceCopyFunc(srcElem, dstElem string) (string, string) {
	funcName := getSliceCopyFuncName(srcElem, dstElem)

	// 如果元素类型相同，直接返回浅拷贝
	if srcElem == dstElem {
		return fmt.Sprintf("func(src []%s) []%s { return append([]%s(nil), src...) }", srcElem, dstElem, dstElem), ""
	}

	// 使用 LoadOrStore 确保并发安全
	if _, loaded := generatedFunctions.Load(funcName); loaded {
		log.Printf("Slice function %s already generated", funcName)
		return funcName, ""
	}

	// 生成基本类型之间的转换函数
	code0, importPath := handleBasicConversion(srcElem, dstElem, true)
	code := fmt.Sprintf(`
    package main
    // %s 是自动生成的切片拷贝函数
    func %s(src []%s) []%s {
        if src == nil {
            return nil
        }
        dst := make([]%s, len(src))
        for i := range src {
            dst[i] = %s(src[i])
        }
        return dst
    }`, funcName, funcName, srcElem, dstElem, dstElem, code0)

	// 安全解析生成的代码
	fset := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		log.Printf("Failed to parse generated slice function: %v", err)
		return "", ""
	}

	if len(parsedFile.Decls) == 0 {
		log.Printf("Generated slice function is empty")
		return "", ""
	}

	if fn, ok := parsedFile.Decls[0].(*ast.FuncDecl); ok {
		addGeneratedFunction(funcName, fn)
		return funcName, importPath
	}
	return "", ""
}

func generateSliceCopyFunc(srcElem, dstElem, elemConv string, file *ast.File, path string) (string, string) {

	// 基本类型直接转换
	if isBasicType(srcElem) && isBasicType(dstElem) {
		return generateBasicSliceCopyFunc(srcElem, dstElem)
	}

	// 强制生成元素类型的转换函数
	generateCopyFunctionIfNeeded(srcElem, dstElem, file, path)

	funcName := getSliceCopyFuncName(srcElem, dstElem)

	// 如果元素类型相同，直接返回浅拷贝
	if srcElem == dstElem {
		return fmt.Sprintf("func(src []%s) []%s { return append([]%s(nil), src...) }", srcElem, dstElem, dstElem), ""
	}
	// 使用 LoadOrStore 确保并发安全
	if _, loaded := generatedFunctions.Load(funcName); loaded {
		log.Printf("Slice function %s already generated", funcName)
		return funcName, ""
	}

	// 处理file为nil的情况
	srcIsStruct := file != nil && isStructType(srcElem, file, path)
	dstIsStruct := file != nil && isStructType(dstElem, file, path)

	// 如果源和目标元素都不是结构体，生成直接拷贝函数
	if !srcIsStruct && !dstIsStruct {
		return fmt.Sprintf("func(src []%s) []%s { return src }", srcElem, dstElem), ""
	}

	// 需要转换函数时，确保elemConv非空
	if elemConv == "" {
		log.Printf("elemConv is required for struct elements but is empty")
		return "", ""
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
		return "", ""
	}

	if len(parsedFile.Decls) == 0 {
		log.Printf("Generated slice function is empty")
		return "", ""
	}

	if fn, ok := parsedFile.Decls[0].(*ast.FuncDecl); ok {
		addGeneratedFunction(funcName, fn)
		return funcName, ""
	}
	return "", ""
}
func handleSliceConversion(srcType, dstType string, allowNarrow, singleToSlice bool, file *ast.File, path string) (string, string) {

	srcElem := getElementType(srcType)
	dstElem := getElementType(dstType)

	if isBasicType(srcElem) && isBasicType(dstElem) {
		return generateBasicSliceCopyFunc(srcElem, dstElem)
	}

	// 生成元素转换函数
	elemConv := getStructCopyFuncName(srcElem, dstElem)
	generateCopyFunctionIfNeeded(srcElem, dstElem, file, path)

	// 只有当元素类型需要转换时才生成切片函数
	if elemConv != "" {
		log.Printf("Generating slice conversion function for %s to %s, funcName: %s", srcType, dstType, elemConv)
		return generateSliceCopyFunc(srcElem, dstElem, elemConv, file, path)
	}
	return "", "" // 直接赋值

}
