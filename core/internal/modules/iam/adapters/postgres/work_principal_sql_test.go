package postgres

import (
	"strings"
	"testing"
)

func TestWorkPrincipalSQLUsesOnlyIAMOwnedRoutines(t *testing.T) {
	for name, statement := range map[string]string{
		"evidence":  workPrincipalEvidenceSQL,
		"directory": workAvailableUsersSQL,
	} {
		if !strings.Contains(statement, "(pg_catalog.current_schemas(false))[2]::name") ||
			strings.Contains(statement, "FROM iam_users") {
			t.Fatalf("%s SQL bypasses IAM routine: %s", name, statement)
		}
	}
	if !strings.Contains(workPrincipalEvidenceSQL, "iam_work_principal_evidence_v1(") ||
		!strings.Contains(workAvailableUsersSQL, "iam_work_available_users_v1(") {
		t.Fatal("Work principal SQL routine binding is incomplete")
	}
}
