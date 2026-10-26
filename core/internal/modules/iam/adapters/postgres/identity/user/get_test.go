package userpostgres

import (
	"strings"
	"testing"
)

func TestGetUserDelegatesReadRoutine(t *testing.T) {
	if !strings.Contains(getUserSQL, "FROM iam_identity_user_read_v3(") ||
		strings.Contains(getUserSQL, "iam_identity_user_read_v1(") ||
		!strings.Contains(getUserSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("get user SQL bypasses the production tenant routine")
	}
}
