package postgres

import (
	"bytes"
	"testing"
)

var iamRBACHTTPMigrationNames = []string{
	"20260716015700_create_iam_rbac_assignment_groups",
	"20260716015701_create_iam_rbac_assignment_members",
	"20260716015702_create_iam_rbac_assignment_helpers_v1",
	"20260716015703_create_iam_rbac_assignment_binding_v1",
	"20260716015704_create_iam_rbac_assignment_create_v1",
	"20260716015705_create_iam_rbac_assignment_update_v1",
	"20260716015706_create_iam_rbac_assignment_retire_v1",
	"20260716015707_create_iam_rbac_assignment_read_v1",
	"20260716015708_create_iam_rbac_effective_read_v1",
	"20260716015709_grant_iam_rbac_assignment_runtime_v1",
}

func TestIAMRBACHTTPMigrationsStayPairedAndBounded(t *testing.T) {
	for _, name := range iamRBACHTTPMigrationNames {
		fresh := readIAMIdentityMigration(t, "fresh", name)
		adopted := readIAMIdentityMigration(t, "v1_adopted", name)
		if !bytes.Equal(fresh, adopted) {
			t.Fatalf("fresh and v1_adopted RBAC HTTP migration differs: %s", name)
		}
		if bytes.Count(fresh, []byte("\n"))+1 > 200 {
			t.Fatalf("RBAC HTTP migration exceeds 200 lines: %s", name)
		}
	}
}
