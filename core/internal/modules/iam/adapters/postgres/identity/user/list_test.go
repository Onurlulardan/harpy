package userpostgres

import (
	"strings"
	"testing"
)

func TestListUsersDelegatesListRoutine(t *testing.T) {
	if !strings.Contains(listUsersSQL, "FROM iam_identity_user_list_v3(") ||
		strings.Contains(listUsersSQL, "iam_identity_user_list_v1(") ||
		!strings.Contains(listUsersSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("list users SQL bypasses the production tenant routine")
	}
}
