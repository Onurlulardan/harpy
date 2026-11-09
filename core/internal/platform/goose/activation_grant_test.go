package goose

import (
	"os"
	"strings"
	"testing"
)

func TestActivationGrantSQLNamesTheTenantRole(t *testing.T) {
	role := "harpy_td_0123456789abcdef0123456789abcdef"
	statements, err := activationGrantSQL(role)
	if err != nil || len(statements) != 4 {
		t.Fatalf("statements=%d err=%v", len(statements), err)
	}
	for _, statement := range statements {
		if !strings.Contains(statement, `"harpy_td_0123456789abcdef0123456789abcdef"`) ||
			strings.Contains(statement, "PUBLIC") {
			t.Fatal(statement)
		}
	}
	joined := strings.Join(statements, "\n")
	for _, signature := range []string{
		"platform.tenancy_activation_assert_v2(uuid, uuid, name)",
		"platform.tenancy_activation_step_context_assert_v1(uuid, text, uuid, name)",
		"platform.tenancy_activation_step_consume_v1(uuid, text, uuid, name)",
		"GRANT USAGE ON SCHEMA platform",
		"WITH GRANT OPTION",
	} {
		if !strings.Contains(joined, signature) {
			t.Fatalf("missing %s", signature)
		}
	}
	if strings.Count(joined, "WITH GRANT OPTION") != 3 {
		t.Fatal(joined)
	}
}

func TestActivationGrantOptionRevokeKeepsExecute(t *testing.T) {
	role := "harpy_td_0123456789abcdef0123456789abcdef"
	statements, err := activationGrantOptionRevokeSQL(role)
	if err != nil || len(statements) != 3 {
		t.Fatalf("statements=%d err=%v", len(statements), err)
	}
	joined := strings.Join(statements, "\n")
	if !strings.Contains(joined, "REVOKE GRANT OPTION FOR EXECUTE") ||
		!strings.Contains(joined, "CASCADE") ||
		strings.Contains(joined, "REVOKE EXECUTE") ||
		strings.Contains(joined, "REVOKE USAGE") ||
		strings.Contains(joined, "REVOKE ALL") {
		t.Fatal(joined)
	}
}

func TestActivationGrantSQLRejectsArbitraryRoles(t *testing.T) {
	if _, err := activationGrantSQL("harpy_owner"); err == nil {
		t.Fatal("owner role was accepted")
	}
	if _, err := activationGrantSQL(`harpy_td_0123456789abcdef0123456789abcdef"; DROP ROLE x`); err == nil {
		t.Fatal("injected role was accepted")
	}
}

func TestInstallationIdentityMigrationStillSelfGrantsPlatformRoutines(t *testing.T) {
	body, err := os.ReadFile("sql/tenant/00650_iam_installation_initialize_persist.sql")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, signature := range []string{
		"GRANT EXECUTE ON FUNCTION platform.tenancy_activation_assert_v2(uuid, uuid, name) TO CURRENT_USER;",
		"GRANT EXECUTE ON FUNCTION platform.tenancy_activation_step_context_assert_v1(uuid, text, uuid, name) TO CURRENT_USER;",
		"GRANT EXECUTE ON FUNCTION platform.tenancy_activation_step_consume_v1(uuid, text, uuid, name) TO CURRENT_USER;",
	} {
		if !strings.Contains(text, signature) {
			t.Fatalf("missing immutable grant %s", signature)
		}
	}
}
