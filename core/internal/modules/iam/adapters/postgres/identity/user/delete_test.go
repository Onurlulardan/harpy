package userpostgres

import (
	"strings"
	"testing"
)

func TestDeleteUserDelegatesTenantAndPlatformRoutines(t *testing.T) {
	if !strings.Contains(deleteUserSQL, "FROM iam_identity_user_delete_v1(") ||
		!strings.Contains(deleteUserSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("delete user SQL bypasses the production tenant routine")
	}
	if !strings.Contains(directoryTombstoneSQL, "SELECT platform.iam_identity_directory_tombstone_v1(") ||
		strings.Contains(directoryTombstoneSQL, "current_schemas") {
		t.Fatal("delete user SQL bypasses the platform directory routine")
	}
}
