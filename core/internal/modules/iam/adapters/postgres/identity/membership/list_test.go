package membershippostgres

import (
	"strings"
	"testing"
)

func TestListMembershipsDelegatesListRoutine(t *testing.T) {
	if !strings.Contains(listMembershipsSQL, "FROM iam_identity_membership_list_v1(") ||
		!strings.Contains(listMembershipsSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("list memberships SQL bypasses the production tenant routine")
	}
}
