package membershippostgres

import (
	"strings"
	"testing"
)

func TestAddMembershipDelegatesAddRoutine(t *testing.T) {
	if !strings.Contains(addMembershipSQL, "FROM iam_identity_membership_add_v1(") ||
		!strings.Contains(addMembershipSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("add membership SQL bypasses the production tenant routine")
	}
}
