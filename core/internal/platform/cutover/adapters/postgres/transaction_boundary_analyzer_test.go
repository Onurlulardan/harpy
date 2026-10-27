package postgres

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strconv"
	"strings"
	"testing"
)

type transactionBoundary struct {
	root    string
	imports map[string]bool
	calls   map[string]bool
}

type boundaryFunction struct {
	declaration *ast.FuncDecl
	file        *ast.File
	filename    string
}

func parseBoundaryPackage(t *testing.T) map[string]*ast.File {
	t.Helper()
	packages, err := parser.ParseDir(token.NewFileSet(), ".", func(info fs.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string]*ast.File)
	for _, parsedPackage := range packages {
		for filename, file := range parsedPackage.Files {
			files[filename] = file
		}
	}
	return files
}

func inspectTransactionBoundary(
	files map[string]*ast.File, spec transactionBoundary,
) []string {
	functions := make(map[string][]boundaryFunction)
	for filename, file := range files {
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if ok && function.Body != nil {
				functions[function.Name.Name] = append(functions[function.Name.Name], boundaryFunction{
					declaration: function, file: file, filename: filename,
				})
			}
		}
	}
	queue := append([]boundaryFunction(nil), functions[spec.root]...)
	if len(queue) == 0 {
		return []string{"missing root " + spec.root}
	}
	visited := make(map[*ast.FuncDecl]bool)
	checkedFiles := make(map[string]bool)
	violations := make([]string, 0)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if visited[current.declaration] {
			continue
		}
		visited[current.declaration] = true
		if !checkedFiles[current.filename] {
			checkedFiles[current.filename] = true
			violations = append(violations, boundaryImportViolations(current, spec.imports)...)
		}
		ast.Inspect(current.declaration.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			name := boundaryCallName(call.Fun)
			if local := functions[name]; len(local) > 0 {
				queue = append(queue, local...)
			} else if !spec.calls[name] {
				violations = append(violations, fmt.Sprintf("%s calls %s", current.declaration.Name, name))
			}
			return true
		})
	}
	return violations
}

func boundaryImportViolations(
	function boundaryFunction, allowed map[string]bool,
) []string {
	violations := make([]string, 0)
	for _, imported := range function.file.Imports {
		path, err := strconv.Unquote(imported.Path.Value)
		if err != nil || !allowed[path] {
			violations = append(violations, fmt.Sprintf("%s imports %s", function.declaration.Name, path))
		}
	}
	return violations
}

func namedBoundaryFunction(file *ast.File, name string) *ast.FuncDecl {
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == name {
			return function
		}
	}
	return nil
}

func boundaryCallName(expression ast.Expr) string {
	switch value := expression.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.SelectorExpr:
		return value.Sel.Name
	default:
		return "unsupported"
	}
}

func stringSet(values ...string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}
