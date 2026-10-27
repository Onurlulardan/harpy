package postgres

import (
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestIAMProfileKeyRepairMigrationIsPairedAndStable(t *testing.T) {
	const name = "20260730120000_repair_iam_profile_keys_and_initialize_v3"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	if len(fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("IAM profile key repair migration is invalid: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"iam_profile_installation_legacy_document_v1",
		"iam_profile_display_name_key_repair_v1",
		"pg_catalog.jsonb_object_keys(p_profile_values)",
		"p_profile_values ? p_stable_field_id::text",
		"p_profile_sha256 IS DISTINCT FROM pg_catalog.sha256",
		"pg_catalog.convert_to(p_profile_values::text, 'UTF8')",
		"pg_catalog.jsonb_typeof(",
		"pg_catalog.to_json(p_stable_field_id::text)::text",
		"pg_catalog.to_json(",
		"repaired.canonical_json::jsonb",
		"pg_catalog.convert_to(repaired.canonical_json, 'UTF8')",
		"snapshot.field_key, snapshot.value_kind",
		"snapshot.cardinality, snapshot.field_state",
		"p_field_key IS DISTINCT FROM 'display_name'",
		"p_value_kind IS DISTINCT FROM 'string'",
		"p_cardinality IS DISTINCT FROM 'single'",
		"p_field_state IS DISTINCT FROM 'active'",
		"profile_sha256 = pg_catalog.sha256",
		"revision = subject.revision + 1",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM profile key repair migration lacks %q", marker)
		}
	}
	if strings.Contains(source, "CREATE OR REPLACE") ||
		strings.Contains(source, "CREATE FUNCTION iam_installation_initialize_v2") {
		t.Fatal("profile repair migration mutates immutable installation v2")
	}
	for _, table := range []string{"iam_users", "iam_organizations", "iam_roles"} {
		if !strings.Contains(source, "UPDATE "+table+" AS subject") {
			t.Fatalf("IAM profile key repair does not cover %s", table)
		}
	}
	if strings.Contains(source,
		"REVOKE EXECUTE ON FUNCTION iam_installation_initialize_v2") {
		t.Fatal("runtime-24 installation compatibility was revoked")
	}
	const successor = "20260730120001_create_iam_installation_initialize_v3"
	v3Fresh := readIAMIdentityMigration(t, "fresh", successor)
	if len(v3Fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	v3Source := string(v3Fresh)
	for _, marker := range []string{
		"iam_installation_initialize_v3(",
		"v_replayed := iam_installation_initialize_v2(",
		"PERFORM iam_profile_display_name_key_repair_v1('display_name')",
		"RETURN v_replayed",
		"GRANT EXECUTE ON FUNCTION iam_installation_initialize_v3",
	} {
		if !strings.Contains(v3Source, marker) {
			t.Fatalf("IAM installation v3 migration lacks %q", marker)
		}
	}
}

func TestIAMInstallationBindingUsesStableProfileKeySuccessor(t *testing.T) {
	if !strings.Contains(installationInitializeSQL, "iam_installation_initialize_v4") ||
		strings.Contains(installationInitializeSQL, "iam_installation_initialize_v2(") {
		t.Fatal("installation binding does not use stable profile key successor")
	}
}

func TestIAMLegacyProfileWriterIsNormalizedAtDatabaseBoundary(t *testing.T) {
	const name = "20260730120002_normalize_iam_legacy_profile_writes_v1"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	if len(fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("IAM legacy profile writer normalizer is invalid: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"iam_profile_display_name_write_normalize_v1()",
		"pg_catalog.jsonb_object_keys(NEW.profile_values)",
		"NEW.profile_values ? 'display_name'",
		"v_display_count IS DISTINCT FROM 1",
		"v_exact_count IS DISTINCT FROM 1",
		"iam_profile_installation_legacy_document_v1(",
		"NEW.profile_values := v_canonical::jsonb",
		"NEW.profile_sha256 := pg_catalog.sha256",
		"BEFORE INSERT OR UPDATE OF profile_values,profile_sha256 ON iam_users",
		"BEFORE INSERT OR UPDATE OF profile_values,profile_sha256 ON iam_organizations",
		"BEFORE INSERT OR UPDATE OF profile_values,profile_sha256 ON iam_roles",
		"ENABLE ALWAYS TRIGGER",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM legacy profile writer normalizer lacks %q", marker)
		}
	}
}
