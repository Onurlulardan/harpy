package userpostgres

import (
	"strings"
	"testing"
)

func TestDisableUserDelegatesDisableRoutine(t *testing.T) {
	if !strings.Contains(disableUserSQL, "FROM iam_identity_user_disable_v1(") ||
		!strings.Contains(disableUserSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("disable user SQL bypasses the production tenant routine")
	}
}
