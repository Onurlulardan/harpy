package migrator

import (
	"errors"
	"strings"
	"testing"
)

func TestMigrationFailureDetailKeepsConstraintAndDropsSQL(t *testing.T) {
	detail := migrationFailureDetail(errors.New(
		`SQL: SELECT 1; ERROR: new row for relation "domain_cutovers" ` +
			`violates check constraint "domain_cutovers_domain_check" (SQLSTATE 23514)`,
	))
	if !strings.Contains(detail, "domain_cutovers_domain_check") ||
		!strings.Contains(detail, "SQLSTATE 23514") ||
		strings.Contains(detail, "SELECT ") {
		t.Fatalf("detail = %q", detail)
	}
}

func TestMigrationFailureDetailRedactsConnectionString(t *testing.T) {
	detail := migrationFailureDetail(errors.New(
		"dial postgres://harpy:secret@127.0.0.1:5432/harpy",
	))
	if detail != "migration execution failed" {
		t.Fatalf("detail = %q", detail)
	}
}
