package userpostgres

import (
	"strings"
	"testing"
)

func TestCreateUserDelegatesTenantAndPlatformRoutines(t *testing.T) {
	if !strings.Contains(createUserSQL, "FROM iam_identity_user_create_v2(") ||
		!strings.Contains(createUserSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("create user SQL bypasses the production tenant routine")
	}
	if !strings.Contains(directoryUpsertSQL, "SELECT platform.iam_identity_directory_upsert_v1(") ||
		strings.Contains(directoryUpsertSQL, "current_schemas") {
		t.Fatal("create user SQL bypasses the platform directory routine")
	}
}
