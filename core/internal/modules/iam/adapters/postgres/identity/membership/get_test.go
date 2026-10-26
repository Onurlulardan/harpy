package membershippostgres

import (
	"strings"
	"testing"
)

func TestGetMembershipDelegatesReadRoutine(t *testing.T) {
	if !strings.Contains(getMembershipSQL, "FROM iam_identity_membership_read_v1(") ||
		!strings.Contains(getMembershipSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("get membership SQL bypasses the production tenant routine")
	}
}
