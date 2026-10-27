package postgres

import (
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestIAMInstallationCapabilityFenceIsPairedAndComplete(t *testing.T) {
	const name = "20260729100035_assert_iam_installation_capabilities_v1"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	if len(fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("IAM installation capability fence is invalid: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"iam_installation_initialize_v2(",
		"iam_installation_credential_initialize_v2(",
		"iam_initial_admin_setup_finalize_v2(",
		"platform.tenancy_activation_step_context_assert_v1(",
		"platform.tenancy_activation_step_consume_v1(",
		"iam_installation_initialize_v1(",
		"iam_installation_credential_initialize_v1(",
		"iam_initial_admin_setup_finalize_v1(",
		"routine.proowner",
		"routine.prosecdef",
		"current_setting('search_path')",
		"'harpy_runtime', routine_oid, 'EXECUTE'",
		"'harpy_v2_runtime', routine_oid, 'EXECUTE'",
		"SELECT 1 / CASE WHEN",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM installation capability fence lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"-- +goose Down", "CREATE OR REPLACE", "DO $",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("IAM installation capability fence contains %q", forbidden)
		}
	}
}

func TestIAMInstallationAuthorityContractsArePairedAndComplete(t *testing.T) {
	tests := []struct {
		scope    string
		name     string
		retired  []string
		retained []string
	}{
		{
			scope: "platform",
			name:  "20260729100038_contract_iam_installation_platform_authority_v1",
			retired: []string{
				"iam_security_epoch_initialize_v1(uuid,timestamptz)",
				"iam_installation_instance_owner_bind_v1(uuid,uuid,uuid,name,uuid,uuid,uuid,uuid,uuid,bigint,text)",
			},
			retained: []string{
				"iam_installation_activation_assert_v2",
				"iam_installation_security_epoch_initialize_v2",
				"iam_installation_identity_directory_upsert_v1",
				"iam_installation_instance_owner_bind_v2",
				"iam_installation_instance_owner_input_v1",
			},
		},
		{
			scope: "tenant",
			name:  "20260729100046_contract_iam_installation_tenant_authority_v1",
			retired: []string{
				"iam_installation_initialize_v1(uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid)",
				"iam_installation_credential_initialize_v1(name,uuid,uuid,uuid,uuid,bigint,bytea,bytea)",
				"iam_initial_admin_setup_finalize_v1(name,uuid,uuid,text)",
			},
			retained: []string{
				"iam_installation_initialize_v2",
				"iam_installation_credential_initialize_v2",
				"iam_initial_admin_setup_finalize_v2",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.scope, func(t *testing.T) {
			fresh := readIAMCapabilityMigration(t, test.scope, "fresh", test.name)
			if len(fresh) == 0 {
				t.Fatal("fresh migration is empty")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("IAM authority contract is invalid: %v", err)
			}
			source := string(fresh)
			for _, marker := range append(test.retired, test.retained...) {
				if !strings.Contains(source, marker) {
					t.Fatalf("IAM authority contract lacks %q", marker)
				}
			}
			for _, marker := range []string{
				"retired_resolved", "retained_resolved", "internal_resolved",
				"routine_oid IS NULL", "routine.proowner", "routine.prosecdef",
				"'harpy_runtime', routine_oid, 'EXECUTE'",
				"'harpy_v2_runtime', routine_oid, 'EXECUTE'",
				"platform.tenancy_activation_step_context_assert_v1",
				"platform.tenancy_activation_step_consume_v1",
			} {
				if !strings.Contains(source, marker) {
					t.Fatalf("IAM authority contract lacks %q", marker)
				}
			}
			for _, forbidden := range []string{
				"-- +goose Down", "CREATE FUNCTION", "CREATE OR REPLACE", "DROP ",
			} {
				if strings.Contains(source, forbidden) {
					t.Fatalf("IAM authority contract contains %q", forbidden)
				}
			}
		})
	}
}
