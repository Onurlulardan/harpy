package cutover

import (
	"path"
	"strings"
	"testing"
)

func TestIAMSystemResourcesFinalizedSwitchMigrationsArePaired(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	versions := []string{"20260716020000", "20260716020001"}
	pairs := make(map[string]map[string]string, len(versions))
	for _, document := range documents {
		name := path.Base(document.Path())
		for _, version := range versions {
			if !strings.HasPrefix(name, version+"_") {
				continue
			}
			origin := "fresh"
			if strings.Contains(name, ".v1_adopted.sql") {
				origin = "v1_adopted"
			}
			if pairs[version] == nil {
				pairs[version] = map[string]string{}
			}
			pairs[version][origin] = string(document.Content())
		}
	}
	for _, version := range versions {
		pair := pairs[version]
		if len(pair) != 2 || pair["fresh"] != pair["v1_adopted"] {
			t.Fatalf("migration %s is not byte-identical and paired", version)
		}
		if strings.Contains(pair["fresh"], "-- +goose Down") {
			t.Fatalf("migration %s has a destructive down section", version)
		}
	}
}

func TestIAMSystemResourcesFinalizedSwitchKeepsFinalizerSurfaceClosed(t *testing.T) {
	sources := iamSystemResourcesSwitchSources(t)
	apply := sources["20260716020000"]
	for _, marker := range []string{
		"SESSION_USER IS DISTINCT FROM 'harpy_cutover_finalizer'",
		`'["iam","system_resource_layout"]'::jsonb`,
		"adoption.state='accepted'", "adoption.item_count=0",
		"FOR SHARE OF adoption,zero_receipt",
		"event_kind,scope_kind,scope_id", "'switch'",
	} {
		if !strings.Contains(apply, marker) {
			t.Fatalf("finalized switch lacks %q", marker)
		}
	}
	attach := sources["20260716020001"]
	for _, marker := range []string{
		"cutover_authority.verify_finalization_receipt_v1",
		"AFTER UPDATE OF state ON platform.cutover_operations",
		"ENABLE ALWAYS TRIGGER",
		"TO harpy_cutover_authority_owner",
	} {
		if !strings.Contains(attach, marker) {
			t.Fatalf("finalizer transition lacks %q", marker)
		}
	}
	if strings.Contains(attach, "TO harpy_cutover_finalizer") {
		t.Fatal("finalized switch opened a second direct finalizer routine")
	}
}

func iamSystemResourcesSwitchSources(t *testing.T) map[string]string {
	t.Helper()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]string{}
	for _, document := range documents {
		if strings.Contains(document.Path(), "/fresh/2026071602000") {
			name := path.Base(document.Path())
			sources[name[:14]] = string(document.Content())
		}
	}
	if len(sources) != 2 {
		t.Fatalf("finalized switch migration count = %d, want 2", len(sources))
	}
	return sources
}
