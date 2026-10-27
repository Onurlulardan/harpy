package postgres

import (
	"strings"
	"testing"
)

func TestIAMRBACMigrationContracts(t *testing.T) {
	tests := []struct {
		name    string
		markers []string
	}{
		{iamRBACMigrationNames[0], []string{
			"CREATE TABLE iam_resources (", "UNIQUE (resource_key)",
			"owner_domain text COLLATE \"C\"", "state IN ('active', 'retired')",
			"CREATE FUNCTION iam_forbid_rbac_mutation_v1()",
			"CREATE FUNCTION iam_guard_rbac_terminal_v1()",
		}},
		{iamRBACMigrationNames[1], []string{
			"CREATE TABLE iam_actions (", "UNIQUE (action_key)",
			"iam_actions_terminal_guard", "iam_actions_removal_guard",
		}},
		{iamRBACMigrationNames[2], []string{
			"CREATE TABLE iam_permissions (", "UNIQUE NULLS NOT DISTINCT",
			"match_kind IN ('exact', 'resource_all_actions', 'tenant_all')",
			"match_kind = 'exact' OR is_protected",
			"FOREIGN KEY (resource_id)", "FOREIGN KEY (action_id)",
			"CREATE FUNCTION iam_guard_active_permission_catalog_v1()",
		}},
		{iamRBACMigrationNames[3], []string{
			"CREATE TABLE iam_user_permission_bindings (",
			"UNIQUE (user_id, permission_id)",
			"CHECK (permission_match_kind = 'exact')",
			"CREATE FUNCTION iam_guard_active_user_binding_v1()",
		}},
		{iamRBACMigrationNames[4], []string{
			"CREATE TABLE iam_role_permission_bindings (",
			"UNIQUE (role_id, permission_id)",
			"permission_match_kind = 'exact' OR role_is_protected",
			"CREATE FUNCTION iam_guard_active_role_binding_v1()",
		}},
		{iamRBACMigrationNames[5], []string{
			"CREATE TABLE iam_organization_permission_bindings (",
			"UNIQUE (organization_id, permission_id)",
			"CHECK (permission_match_kind = 'exact')",
			"CREATE FUNCTION iam_guard_active_organization_binding_v1()",
		}},
		{iamRBACMigrationNames[6], []string{
			"CREATE TABLE iam_permission_implications (",
			"implication_kind = 'v1_form_parent_v1'",
			"source_match_kind = 'exact' AND implied_match_kind = 'exact'",
			"hit_count >= 0", "retirement_receipt_id IS NOT NULL",
			"CREATE FUNCTION iam_guard_permission_implication_v1()",
		}},
		{iamRBACMigrationNames[7], []string{
			"CREATE TABLE iam_authorization_clock (",
			"CREATE TABLE iam_authorization_operation_receipts (",
			"PRIMARY KEY (operation_id)", "pg_catalog.octet_length(request_sha256) = 32",
			"NEW.clock_value IS DISTINCT FROM OLD.clock_value + 1",
		}},
		{iamRBACMigrationNames[8], []string{
			"CREATE TABLE iam_protected_admin_policies (",
			"permission_match_kind = 'tenant_all'",
			"minimum_active_assignments >= 1", "role_scope_kind = 'tenant'",
			"CREATE FUNCTION iam_guard_protected_admin_policy_v1()",
		}},
		{iamRBACMigrationNames[9], []string{
			"CREATE FUNCTION iam_guard_role_wildcard_policy_v1()",
			"CREATE FUNCTION iam_guard_protected_admin_minimum_v1()",
			"DEFERRABLE INITIALLY DEFERRED",
			"iam_users_protected_minimum_guard",
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := string(readIAMIdentityMigration(t, "fresh", test.name))
			for _, marker := range test.markers {
				if !strings.Contains(source, marker) {
					t.Fatalf("IAM RBAC migration lacks exact marker %q", marker)
				}
			}
		})
	}
}

func TestIAMRBACUsesOnlyReservedMigrationVersions(t *testing.T) {
	want := []string{
		"20260716013020", "20260716013030", "20260716013040", "20260716013050",
		"20260716013100", "20260716013110", "20260716013120", "20260716013130",
		"20260716013140", "20260716013150",
	}
	for index, name := range iamRBACMigrationNames {
		if !strings.HasPrefix(name, want[index]+"_") {
			t.Fatalf("RBAC migration %q does not use reserved version %s", name, want[index])
		}
	}
}
