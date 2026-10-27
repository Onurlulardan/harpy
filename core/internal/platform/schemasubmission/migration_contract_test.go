package schemasubmission_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSchemaSubmissionMigrationPairsAreExactAndOwnerOnly(t *testing.T) {
	root := repositoryRoot(t)
	versions := []string{
		"20260718020000_add_schema_submission_canonical_bytes",
		"20260718020001_create_schema_submission_step_validator_v1",
		"20260718020002_create_schema_submission_assertion_v1",
		"20260718020003_create_schema_submission_replay_v1",
		"20260718020004_create_schema_plan_submission_v1",
		"20260718020005_create_schema_result_resolver_v1",
	}
	var combined strings.Builder
	for _, version := range versions {
		fresh := readMigration(t, root, "fresh", version+".fresh.sql")
		adopted := readMigration(t, root, "v1_adopted", version+".v1_adopted.sql")
		if !bytes.Equal(fresh, adopted) {
			t.Fatalf("origin pair differs for %s", version)
		}
		combined.Write(fresh)
	}
	document := combined.String()
	for _, required := range []string{
		"canonical_plan_bytes bytea",
		"canonical_step_bytes bytea",
		"platform_migrator_submit_schema_plan_v1",
		"platform_migrator_resolve_schema_result_v1",
		"platform.billing_resolve_current_tenant_authority_v1",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"pg_catalog.uuid_extract_version",
		"pg_catalog.sha256",
		"SCHEMA_JOB_FAILED",
		"REVOKE ALL ON FUNCTION",
	} {
		if !strings.Contains(document, required) {
			t.Fatalf("migration contract lacks %q", required)
		}
	}
	if strings.Contains(document, "GRANT EXECUTE") {
		t.Fatal("owner-only helper unexpectedly grants execute")
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "../../../.."))
}

func readMigration(t *testing.T, root, variant, name string) []byte {
	t.Helper()
	path := filepath.Join(root, "core", "internal", "platform", "migrator",
		"migrations", variant, name)
	document, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return document
}
