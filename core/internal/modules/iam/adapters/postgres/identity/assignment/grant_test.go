package assignmentpostgres

import (
	"strings"
	"testing"
)

func TestGrantAssignmentDelegatesGrantRoutine(t *testing.T) {
	if !strings.Contains(grantAssignmentSQL, "FROM iam_identity_assignment_grant_v1(") ||
		!strings.Contains(grantAssignmentSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("grant assignment SQL bypasses the production tenant routine")
	}
}
