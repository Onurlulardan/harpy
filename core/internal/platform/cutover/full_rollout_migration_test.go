package cutover

import (
	"path"
	"strings"
	"testing"
)

func TestFullRolloutCompletionMigrationsArePairedAndForwardOnly(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	versions := []string{
		"20260718230000", "20260718230001", "20260718230002", "20260718230003",
	}
	fresh := make(map[string]string, len(versions))
	for _, document := range documents {
		name := path.Base(document.Path())
		if !strings.HasSuffix(name, ".fresh.sql") {
			continue
		}
		for _, version := range versions {
			if strings.HasPrefix(name, version+"_") {
				fresh[version] = string(document.Content())
			}
		}
	}
	for _, version := range versions {
		source := fresh[version]
		if source == "" {
			t.Fatalf("fresh full rollout migration %s did not load", version)
		}
		if strings.Count(source, "-- +goose Up") != 1 ||
			strings.Contains(source, "-- +goose Down") {
			t.Fatalf("full rollout migration %s is not one forward-only Goose block", version)
		}
	}
}

func TestFullRolloutCompletionMigrationsKeepProtectedSurfaceClosed(t *testing.T) {
	sources := fullRolloutMigrationSources(t)
	table := sources["20260718230000"]
	for _, marker := range []string{
		"UNIQUE (environment_id, tenant_id)",
		"full_rollout_completion_zero_v1_check",
		"last_legacy_hit_at <= recorded_at - interval '30 days'",
		"full_rollout_completion_receipts_immutable",
		"REVOKE ALL ON ALL TABLES IN SCHEMA cutover_authority",
	} {
		if !strings.Contains(table, marker) {
			t.Fatalf("full rollout receipt table lacks %q", marker)
		}
	}
	routine := sources["20260718230003"]
	for _, marker := range []string{
		"rollback_window_closure_sha256 bytea NOT NULL",
		"p_rollback_window_closure_sha256",
		"SESSION_USER IS DISTINCT FROM 'harpy_cluster_admin'",
		"p_rollback_window_closed_at < p_last_legacy_hit_at",
		"FOR SHARE;",
		`'["billing"]'::jsonb`,
		`'["iam","system_resource_layout"]'::jsonb`,
		"retirement.state IS DISTINCT FROM 'retired'",
		"ON CONFLICT (receipt_id) DO NOTHING",
		"FROM PUBLIC;",
	} {
		if !strings.Contains(routine, marker) {
			t.Fatalf("protected full rollout routine lacks %q", marker)
		}
	}
	grant := sources["20260718230003"]
	if !strings.Contains(grant, "DROP FUNCTION") ||
		!strings.Contains(grant, ") FROM harpy_cluster_admin;") ||
		!strings.Contains(grant, ") TO harpy_cluster_admin;") ||
		strings.Contains(grant, "TO harpy_runtime") ||
		strings.Contains(grant, "TO harpy_cutover_finalizer") {
		t.Fatal("full rollout grant opened a non-cluster-admin caller")
	}
}

func fullRolloutMigrationSources(t *testing.T) map[string]string {
	t.Helper()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]string{}
	for _, document := range documents {
		if strings.Contains(document.Path(), "/fresh/2026071823000") {
			name := path.Base(document.Path())
			sources[name[:14]] = string(document.Content())
		}
	}
	if len(sources) != 4 {
		t.Fatalf("full rollout migration count = %d, want 4", len(sources))
	}
	return sources
}
