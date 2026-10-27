package membershippostgres

import (
	"strings"
	"testing"
)

func TestUpdateMembershipDelegatesUpdateRoutine(t *testing.T) {
	if !strings.Contains(updateMembershipSQL, "FROM iam_identity_membership_update_v1(") ||
		!strings.Contains(updateMembershipSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("update membership SQL bypasses the production tenant routine")
	}
}
