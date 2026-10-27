package assignmentpostgres

import (
	"strings"
	"testing"
)

func TestGetAssignmentDelegatesReadRoutine(t *testing.T) {
	if !strings.Contains(getAssignmentSQL, "FROM iam_identity_assignment_read_v1(") ||
		!strings.Contains(getAssignmentSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("get assignment SQL bypasses the production tenant routine")
	}
}
