package postgres

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"testing"
)

func TestPublicAPIDoesNotExposeRawSchemaExecutor(t *testing.T) {
	packages, err := parser.ParseDir(
		token.NewFileSet(), ".",
		func(info fs.FileInfo) bool {
			return info.Name() != "" && !hasTestSuffix(info.Name())
		},
		0,
	)
	if err != nil {
		t.Fatal(err)
	}
	parsed := packages["postgres"]
	if parsed == nil {
		t.Fatal("postgres package was not parsed")
	}
	for _, file := range parsed.Files {
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || !function.Name.IsExported() {
				continue
			}
			if function.Name.Name == "AcquireSchemaExecutor" {
				t.Fatal("raw schema executor acquisition API is forbidden")
			}
			if publicRawMethod(function) {
				t.Fatalf("raw PostgreSQL method %s is publicly exposed", function.Name.Name)
			}
		}
	}
}

func publicRawMethod(function *ast.FuncDecl) bool {
	if function.Recv == nil || len(function.Recv.List) != 1 {
		return false
	}
	receiver := receiverIdentifier(function.Recv.List[0].Type)
	if receiver == nil || !receiver.IsExported() || receiver.Name == "BoundExecutor" {
		return false
	}
	switch function.Name.Name {
	case "Exec", "Query", "QueryRow":
		return true
	default:
		return false
	}
}

func receiverIdentifier(expression ast.Expr) *ast.Ident {
	if pointer, ok := expression.(*ast.StarExpr); ok {
		expression = pointer.X
	}
	receiver, _ := expression.(*ast.Ident)
	return receiver
}

func hasTestSuffix(name string) bool {
	const suffix = "_test.go"
	return len(name) >= len(suffix) && name[len(name)-len(suffix):] == suffix
}
