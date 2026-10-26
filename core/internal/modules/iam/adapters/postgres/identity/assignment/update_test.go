package assignmentpostgres

import (
	"strings"
	"testing"
)

func TestUpdateAssignmentDelegatesUpdateRoutine(t *testing.T) {
	if !strings.Contains(updateAssignmentSQL, "FROM iam_identity_assignment_update_v1(") ||
		!strings.Contains(updateAssignmentSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("update assignment SQL bypasses the production tenant routine")
	}
}
