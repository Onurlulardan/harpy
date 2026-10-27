package postgres

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestTargetTransactionsHaveNoExternalEffectSurface(t *testing.T) {
	files := parseBoundaryPackage(t)
	tests := []struct {
		name string
		spec transactionBoundary
	}{
		{name: "apply", spec: transactionBoundary{
			root: "applyTargetTransaction",
			imports: stringSet(
				"context", "embed", "encoding/json", "fmt", "time",
				"github.com/Onurlulardan/harpy/contracts/cutover/v1",
				"github.com/Onurlulardan/harpy/core/internal/platform/cutover",
				"github.com/jackc/pgx/v5", "github.com/jackc/pgx/v5/pgtype",
			),
			calls: stringSet(
				"Background", "BeginTx", "Commit", "Errorf", "Lookup", "QueryRow", "RawMessage",
				"Rollback", "Scan", "ValidateApplyResult", "WithTimeout", "cancel", "int64",
				"len", "uint32",
			),
		}},
		{name: "finalization", spec: transactionBoundary{
			root: "finalizeTargetTransaction",
			imports: stringSet(
				"context", "fmt", "time", "github.com/jackc/pgx/v5",
			),
			calls: stringSet(
				"Background", "BeginTx", "Commit", "Errorf", "QueryRow",
				"Rollback", "Scan", "WithTimeout", "cancel",
			),
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if violations := inspectTransactionBoundary(files, test.spec); len(violations) != 0 {
				t.Fatalf("transaction effect boundary drift: %v", violations)
			}
		})
	}
}

func TestTransactionBoundaryRejectsTransitiveEffectHelpers(t *testing.T) {
	fixtures := []struct {
		name, source string
	}{
		{"network", `package fixture
import "net/http"
func transaction() { wrapper() }
func wrapper() { _, _ = http.Get("https://invalid.example") }
`},
		{"telemetry", `package fixture
import "go.opentelemetry.io/otel"
func transaction() { wrapper() }
func wrapper() { _ = otel.Tracer("cutover") }
`},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), fixture.name+".go", fixture.source, 0)
			if err != nil {
				t.Fatal(err)
			}
			violations := inspectTransactionBoundary(
				map[string]*ast.File{fixture.name + ".go": file},
				transactionBoundary{root: "transaction", imports: stringSet(), calls: stringSet()},
			)
			if len(violations) == 0 {
				t.Fatal("transitive external effect helper was not rejected")
			}
		})
	}
}

func TestFinalizationDocumentsStayBeforeTargetTransaction(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "finalization_database.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	function := namedBoundaryFunction(file, "Finalize")
	if function == nil {
		t.Fatal("Finalize is missing")
	}
	positions := map[string]token.Pos{}
	ast.Inspect(function.Body, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			positions[boundaryCallName(call.Fun)] = call.Pos()
		}
		return true
	})
	if positions["finalizationDocuments"] == 0 || positions["finalizeTargetTransaction"] == 0 ||
		positions["finalizationDocuments"] >= positions["finalizeTargetTransaction"] {
		t.Fatal("verified evidence documents must be prepared before the target transaction")
	}
	for _, forbidden := range []string{"BeginTx", "Query", "QueryRow", "Exec", "Commit"} {
		if positions[forbidden] != 0 {
			t.Fatalf("Finalize performs target transaction call %q directly", forbidden)
		}
	}
}
