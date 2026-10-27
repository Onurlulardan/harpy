package assignmentpostgres

import (
	"strings"
	"testing"
)

func TestRevokeAssignmentDelegatesRevokeRoutine(t *testing.T) {
	if !strings.Contains(revokeAssignmentSQL, "FROM iam_identity_assignment_revoke_v1(") ||
		!strings.Contains(revokeAssignmentSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("revoke assignment SQL bypasses the production tenant routine")
	}
}
