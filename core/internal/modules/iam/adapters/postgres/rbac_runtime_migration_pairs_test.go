package postgres

import (
	"bytes"
	"testing"
)

var iamRBACRuntimeMigrationNames = []string{
	"20260716014000_create_iam_rbac_change_audit",
	"20260716014001_create_iam_rbac_actor_assert_v1",
	"20260716014002_create_iam_rbac_mutation_begin_v1",
	"20260716014003_create_iam_rbac_mutation_finish_v1",
	"20260716014004_create_iam_rbac_user_bumps_v1",
	"20260716014005_create_iam_rbac_read_assert_v1",
	"20260716014010_create_iam_rbac_resource_create_v1",
	"20260716014011_create_iam_rbac_resource_read_v1",
	"20260716014012_create_iam_rbac_resource_update_v1",
	"20260716014013_create_iam_rbac_resource_retire_v1",
	"20260716014020_create_iam_rbac_action_create_v1",
	"20260716014021_create_iam_rbac_action_read_v1",
	"20260716014022_create_iam_rbac_action_update_v1",
	"20260716014023_create_iam_rbac_action_retire_v1",
	"20260716014030_create_iam_rbac_permission_create_v1",
	"20260716014031_create_iam_rbac_permission_read_v1",
	"20260716014032_create_iam_rbac_permission_retire_v1",
	"20260716014040_create_iam_rbac_user_binding_write_v1",
	"20260716014041_create_iam_rbac_user_binding_read_v1",
	"20260716014042_create_iam_rbac_role_binding_write_v1",
	"20260716014043_create_iam_rbac_role_binding_read_v1",
	"20260716014044_create_iam_rbac_organization_binding_write_v1",
	"20260716014045_create_iam_rbac_organization_binding_read_v1",
	"20260716014050_create_iam_rbac_implication_create_v1",
	"20260716014051_create_iam_rbac_implication_read_v1",
	"20260716014052_create_iam_rbac_implication_record_hit_v1",
	"20260716014053_create_iam_rbac_implication_retire_v1",
	"20260716014054_create_iam_rbac_protected_admin_read_v1",
	"20260716014055_create_iam_rbac_protected_admin_update_v1",
	"20260716014056_create_iam_rbac_audit_write_v1",
	"20260716014057_grant_iam_rbac_runtime_v1",
}

func TestIAMRBACRuntimeMigrationsStayPaired(t *testing.T) {
	for _, name := range iamRBACRuntimeMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			if len(fresh) == 0 {
				t.Fatal("fresh migration is empty")
			}
			if bytes.Count(fresh, []byte("\n"))+1 > 200 {
				t.Fatal("RBAC runtime migration exceeds 200 lines")
			}
		})
	}
}

func TestIAMRBACRuntimeMigrationRangeIsClosed(t *testing.T) {
	if len(iamRBACRuntimeMigrationNames) != 31 {
		t.Fatalf("RBAC runtime migration count = %d, want 31", len(iamRBACRuntimeMigrationNames))
	}
}
