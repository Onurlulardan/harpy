package userpostgres

import (
	"strings"
	"testing"
)

func TestUpdateUserDelegatesUpdateRoutine(t *testing.T) {
	if !strings.Contains(updateUserSQL, "FROM iam_identity_user_update_v2(") ||
		!strings.Contains(updateUserSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("update user SQL bypasses the production tenant routine")
	}
}
