package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestIAMInstallationCapabilityMigrationsStayPairedAndAdditive(t *testing.T) {
	tests := []struct {
		scope, name string
		markers     []string
		legacyACL   string
	}{
		{
			scope: "platform",
			name:  "20260729100024_create_iam_installation_capability_entrypoints_v1",
			markers: []string{
				"iam_installation_activation_assert_v2",
				"iam_installation_security_epoch_initialize_v2",
				"iam_installation_identity_directory_upsert_v1",
				"iam_installation_instance_owner_bind_v2",
				"'platformauthority.bind_owner'",
				"tenancy_activation_step_context_assert_v1",
			},
			legacyACL: "REVOKE EXECUTE ON FUNCTION platform.iam_security_epoch_initialize_v1",
		},
		{
			scope: "tenant",
			name:  "20260729100025_create_iam_installation_tenant_capability_entrypoints_v1",
			markers: []string{
				"iam_installation_initialize_v2",
				"iam_installation_credential_initialize_v2",
				"iam_initial_admin_setup_finalize_v2",
				"'environment_initialize'",
				"'saas_customer_tenant_provision'",
				"tenancy_activation_step_consume_v1",
			},
			legacyACL: "REVOKE EXECUTE ON FUNCTION iam_installation_initialize_v1",
		},
	}
	for _, test := range tests {
		t.Run(test.scope, func(t *testing.T) {
			fresh := readIAMCapabilityMigration(t, test.scope, "fresh", test.name)
			adopted := readIAMCapabilityMigration(t, test.scope, "v1_adopted", test.name)
			if !bytes.Equal(fresh, adopted) {
				t.Fatal("fresh and v1_adopted capability migrations differ")
			}
			if bytes.Count(fresh, []byte("\n"))+1 > 200 {
				t.Fatal("IAM installation capability migration exceeds 200 lines")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("invalid IAM installation capability migration: %v", err)
			}
			source := string(fresh)
			for _, marker := range test.markers {
				if !strings.Contains(source, marker) {
					t.Fatalf("capability migration lacks marker %q", marker)
				}
			}
			if strings.Contains(source, test.legacyACL) {
				t.Fatalf("Gen70 migration contracts legacy routine: %s", test.legacyACL)
			}
		})
	}
}

func TestIAMInstallationBindingUsesCapabilitySuccessors(t *testing.T) {
	queries := map[string]string{
		"activation": installationActivationAssertSQL,
		"epoch":      installationSecurityEpochSQL,
		"identity":   installationInitializeSQL,
		"directory":  installationDirectorySQL,
		"owner":      installationPlatformOwnerSQL,
	}
	for name, query := range queries {
		if !strings.Contains(query, "installation") ||
			!strings.Contains(query, "_v2") &&
				!strings.Contains(query, "_v3") &&
				!strings.Contains(query, "_v4") &&
				!strings.Contains(query, "directory_upsert_v1") {
			t.Fatalf("%s binding does not use an installation capability successor", name)
		}
	}
}

func readIAMCapabilityMigration(
	t *testing.T, scope, origin, name string,
) []byte {
	t.Helper()
	if scope == "platform" {
		return readIAMPlatformAuthMigration(t, origin, name)
	}
	return readIAMIdentityMigration(t, origin, name)
}
