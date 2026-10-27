package membershippostgres

import (
	"strings"
	"testing"
)

func TestDeleteMembershipDelegatesDeleteRoutine(t *testing.T) {
	if !strings.Contains(deleteMembershipSQL, "FROM iam_identity_membership_delete_v1(") ||
		!strings.Contains(deleteMembershipSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("delete membership SQL bypasses the production tenant routine")
	}
}
