package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamRBACMigrationNames = []string{
	"20260716013020_create_iam_resources",
	"20260716013030_create_iam_actions",
	"20260716013040_create_iam_permissions",
	"20260716013050_create_iam_user_permission_bindings",
	"20260716013100_create_iam_role_permission_bindings",
	"20260716013110_create_iam_organization_permission_bindings",
	"20260716013120_create_iam_permission_implications",
	"20260716013130_create_iam_authorization_state",
	"20260716013140_create_iam_protected_admin_policy",
	"20260716013150_guard_iam_protected_admin_v1",
}

func TestIAMRBACMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range iamRBACMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			if len(fresh) == 0 {
				t.Fatal("fresh migration is empty")
			}
			assertIAMRBACMigrationSource(t, fresh)
		})
	}
}

func assertIAMRBACMigrationSource(t *testing.T, content []byte) {
	t.Helper()
	if bytes.Count(content, []byte("\n"))+1 > 200 {
		t.Fatal("IAM RBAC migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(content); err != nil {
		t.Fatalf("IAM RBAC migration source is invalid: %v", err)
	}
	upper := strings.ToUpper(string(content))
	for _, forbidden := range []string{
		"IF NOT EXISTS", "-- +GOOSE DOWN", "CREATE OR REPLACE",
		"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP",
		"PUBLIC.", "\nGRANT ", "ON DELETE CASCADE", "ON DELETE SET NULL",
		"HARPY_RUNTIME", "HARPY_ADOPTER", "TENANT_ID",
		"USERS_FORM_TABLE", "ROLES_FORM_TABLE", "ORGANIZATIONS_FORM_TABLE",
	} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("IAM RBAC migration contains forbidden token %q", forbidden)
		}
	}
	functions := strings.Count(upper, "CREATE FUNCTION ")
	if functions > 0 {
		if strings.Count(upper, "SECURITY DEFINER") != functions ||
			strings.Count(upper, "REVOKE ALL ON FUNCTION") != functions {
			t.Fatal("IAM RBAC functions are not closed SECURITY DEFINER routines")
		}
		if !strings.Contains(upper, "FROM PUBLIC") {
			t.Fatal("IAM RBAC function does not revoke PUBLIC execute")
		}
	}
	triggers := strings.Count(upper, "CREATE TRIGGER ") +
		strings.Count(upper, "CREATE CONSTRAINT TRIGGER ")
	if strings.Count(upper, "ENABLE ALWAYS TRIGGER") != triggers {
		t.Fatal("IAM RBAC triggers are not all enabled ALWAYS")
	}
}
