package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestIAMInstallationPlatformOwnerMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range []string{
		"20260725125200_create_iam_installation_instance_owner_bind_v1",
		"20260725125201_create_iam_installation_instance_owner_input_v1",
		"20260725125210_repair_iam_installation_owner_status_case",
		"20260729100048_create_iam_installation_instance_owner_verify_v1",
	} {
		fresh := readIAMPlatformAuthMigration(t, "fresh", name)
		adopted := readIAMPlatformAuthMigration(t, "v1_adopted", name)
		if !bytes.Equal(fresh, adopted) {
			t.Fatalf("%s origin pair differs", name)
		}
		if bytes.Count(fresh, []byte("\n"))+1 > 200 {
			t.Fatalf("%s exceeds 200 lines", name)
		}
		if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
			t.Fatalf("%s is invalid: %v", name, err)
		}
	}
	hotfix := string(readIAMPlatformAuthMigration(t, "fresh",
		"20260725125210_repair_iam_installation_owner_status_case"))
	if !strings.Contains(hotfix, "tenant.status='ACTIVE'") ||
		strings.Contains(hotfix, "tenant.status='active'") {
		t.Fatal("installation Instance Owner status case is not authoritative")
	}
	source := string(readIAMPlatformAuthMigration(t, "fresh",
		"20260725125200_create_iam_installation_instance_owner_bind_v1"))
	for _, marker := range []string{
		"iam_installation_instance_owner_bind_v1",
		"operation.state IN ('finalizing','succeeded')",
		"platform.platformauthority_installation_owner_bindings",
		"platform.iam_login_directory", "'initial_owner'",
		"'instance_owner'", "'environment_all'",
		"platform.iam_platform_authorization_clocks",
		"TO harpy_runtime",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("installation Instance Owner migration lacks %q", marker)
		}
	}
}

func TestIAMInstallationPlatformOwnerVerificationIsReadOnly(t *testing.T) {
	source := string(readIAMPlatformAuthMigration(t, "fresh",
		"20260729100048_create_iam_installation_instance_owner_verify_v1"))
	for _, marker := range []string{
		"iam_installation_instance_owner_verify_v1",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"CURRENT_USER IS DISTINCT FROM 'harpy_owner'",
		"harpy-v2:core-operator-initialize",
		"platform.iam_platform_mutation_receipts",
		"platform.iam_platform_owner_policies",
		"platform.iam_platform_authorization_clocks",
		"'initial_owner'", "'instance_owner'", "'environment_all'",
		"initial_binding.state='active'",
		"initial_binding.state='revoked'",
		"duplicate_receipt", "unexpected_binding",
		"ERRCODE='42501'", "ERRCODE='55000'",
		"FROM PUBLIC", "TO harpy_runtime",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("installation Instance Owner verifier lacks %q", marker)
		}
	}
	upper := strings.ToUpper(source)
	for _, forbidden := range []string{
		"\nINSERT ", "\nUPDATE ", "\nDELETE ", "\nMERGE ",
		"IAM_INSTALLATION_INSTANCE_OWNER_BIND_V1(",
		"IAM_INSTALLATION_INSTANCE_OWNER_BIND_V2(",
		"TENANCY_ACTIVATION_STEP_CONTEXT_ASSERT_V1(",
		"TENANCY_ACTIVATION_STEP_CONSUME_V1(",
	} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("installation Instance Owner verifier contains %q", forbidden)
		}
	}
}
