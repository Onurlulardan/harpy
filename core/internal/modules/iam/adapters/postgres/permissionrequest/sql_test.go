package postgres

import (
	"strings"
	"testing"
)

func TestPermissionRequestSQLUsesPGFirstTenantAndClosedRoutines(t *testing.T) {
	documents := map[string]string{
		"create": permissionRequestCreateSQL, "read": permissionRequestReadSQL,
		"list": permissionRequestListSQL, "review": permissionRequestReviewSQL,
		"cancel": permissionRequestCancelSQL, "expire": permissionRequestExpireSQL,
		"decide": permissionRequestDecideSQL,
	}
	for name, document := range documents {
		if !strings.Contains(document, "(pg_catalog.current_schemas(false))[2]::name") {
			t.Fatalf("%s SQL does not resolve the PG-first tenant schema", name)
		}
		if strings.Contains(document, "pg_catalog.current_schema()") {
			t.Fatalf("%s SQL regressed to current_schema", name)
		}
	}
	for name, document := range map[string]string{
		"read": permissionRequestReadSQL, "list": permissionRequestListSQL,
	} {
		if strings.Contains(document, "iam_permission_request_groups") ||
			strings.Contains(document, "iam_permission_requests AS") {
			t.Fatalf("%s SQL opened permission request tables directly", name)
		}
	}
}
