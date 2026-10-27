package assignmentpostgres

import (
	"strings"
	"testing"
)

func TestListAssignmentsDelegatesListRoutine(t *testing.T) {
	if !strings.Contains(listAssignmentsSQL, "FROM iam_identity_assignment_list_v1(") ||
		!strings.Contains(listAssignmentsSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("list assignments SQL bypasses the production tenant routine")
	}
}
