package goose

import (
	"os"
	"strings"
	"testing"
)

func TestWorkerSchemaContractsStayIdle(t *testing.T) {
	t.Parallel()
	authority := readMigration(t, "sql/platform/00805_platform_cutover_authority.sql")
	up, _, ok := strings.Cut(authority, "-- +goose Down")
	if !ok {
		t.Fatal("cutover migration has no down section")
	}
	for _, want := range []string{
		"CREATE TABLE platform.cutover_groups",
		"CREATE TABLE platform.cutover_group_members",
		"CREATE TABLE platform.domain_cutovers",
		"CREATE TABLE platform.cutover_job_bindings",
		"platform_async_cutover_shadow_operation_active_v1",
		"ADD COLUMN created_at",
		"ADD COLUMN parked_at",
		"GRANT SELECT",
	} {
		if !strings.Contains(up, want) {
			t.Fatalf("cutover migration missing %s", want)
		}
	}

	candidates := readMigration(t, "sql/platform/00806_platform_async_idle_candidates.sql")
	candidateUp, _, ok := strings.Cut(candidates, "-- +goose Down")
	if !ok || strings.Count(candidateUp, "WHERE false") != 5 {
		t.Fatal("idle candidate migration must return no row from five readers")
	}
	if !strings.Contains(candidateUp, "tenant_schema text") {
		t.Fatal("import cleanup reader still omits tenant_schema")
	}

	appliedCleanup := readMigration(t,
		"sql/platform/00311_platform_async_iam_import_cleanup_tenant_v1.sql")
	appliedCandidate := readMigration(t,
		"sql/platform/00680_platform_async_analytics_dashboard_export_candidate_v1.sql")
	if strings.Contains(appliedCleanup, "tenant_schema text") ||
		strings.Contains(appliedCandidate, "WHERE false") {
		t.Fatal("applied stub migration changed")
	}
}

func readMigration(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}
