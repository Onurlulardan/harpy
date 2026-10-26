package postgres

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"testing"
)

func TestRuntimeTransactionIsTheOnlyRuntimeTransactionType(t *testing.T) {
	packages, err := parser.ParseDir(
		token.NewFileSet(), ".",
		func(info fs.FileInfo) bool { return !hasTestSuffix(info.Name()) },
		0,
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range packages["postgres"].Files {
		ast.Inspect(file, func(node ast.Node) bool {
			identifier, ok := node.(*ast.Ident)
			if ok && (identifier.Name == "TenantTransaction" ||
				identifier.Name == "BeginTenantTransaction") {
				t.Fatalf("legacy tenant transaction ABI remains: %s", identifier.Name)
			}
			return true
		})
	}
}
