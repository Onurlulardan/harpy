package migrationbundle

import (
	"bytes"
	"strings"
	"testing"
)

func TestSelectKeepsTerminalSupersession(t *testing.T) {
	const version = "20260730120104_repair_installation_layout_publications_v1.fresh.sql"
	root := "core/internal/modules/system_resource_layout/migrations/tenant/"
	paths := []string{
		root + "fresh/" + version,
		root + "supersessions/fresh/" + version,
		root + "supersessions/v2/fresh/" + version,
		root + "supersessions/v3/fresh/" + version,
	}
	sources := make([]Source, len(paths))
	for index, documentPath := range paths {
		sources[index] = Source{
			Owner: "system_resource_layout", Scope: "tenant", Path: documentPath,
			Body: goose("SELECT 1;"),
		}
	}
	selected, err := Select(sources)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 1 || !strings.Contains(selected[0].Path, "/supersessions/v3/fresh/") {
		t.Fatalf("terminal supersession = %#v", selected)
	}
}

func TestSelectRejectsAdoptedAndAmbiguousCopies(t *testing.T) {
	adopted := Source{
		Owner: "platform_migrator", Scope: "platform",
		Path: "core/internal/platform/migrator/migrations/v1_adopted/" +
			"20260713010000_create_foundation.v1_adopted.sql",
		Body: goose("SELECT 1;"),
	}
	if _, err := Select([]Source{adopted}); err == nil ||
		!strings.Contains(err.Error(), "v1_adopted is not present") {
		t.Fatalf("adopted source error = %v", err)
	}
	path := "core/internal/modules/audit/migrations/tenant/fresh/" +
		"20260801040000_create_audit_change_log.fresh.sql"
	first := Source{Owner: "audit", Scope: "tenant", Path: path, Body: goose("SELECT 1;")}
	second := first
	if _, err := Select([]Source{first, second}); err == nil ||
		!strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("duplicate path error = %v", err)
	}
}

func TestBuildIsDeterministicAndDigestTracksBytes(t *testing.T) {
	source := Source{
		Owner: "audit", Scope: "tenant",
		Path: "core/internal/modules/audit/migrations/tenant/fresh/" +
			"20260801040000_create_audit_change_log.fresh.sql",
		Body: goose("SELECT 1;"),
	}
	_, first, err := Build([]Source{source})
	if err != nil {
		t.Fatal(err)
	}
	_, second, err := Build([]Source{source})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) || Digest(first) != Digest(second) {
		t.Fatal("fresh manifest bytes differ across builds")
	}
	changed := source
	changed.Body = goose("SELECT 2;")
	_, other, err := Build([]Source{changed})
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(first, other) || Digest(first) == Digest(other) {
		t.Fatal("fresh manifest digest ignored SQL bytes")
	}
	if !bytes.Contains(first, []byte(`"phase":"expand"`)) ||
		!bytes.Contains(first, []byte(`"target_database_generation":1`)) ||
		!bytes.Contains(first, []byte(`"next_minimum_runtime_generation":1`)) {
		t.Fatalf("authoring metadata missing: %s", first)
	}
}

func goose(statement string) []byte {
	return []byte("-- +goose Up\n" + statement + "\n")
}
