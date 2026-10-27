package cutover

import (
	"path"
	"strings"
	"testing"
)

func TestSystemResourceLayoutDomainMigrationsArePaired(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	versions := []string{
		"20260716005800", "20260716005810", "20260716005820",
		"20260716005830", "20260716005840", "20260716005850",
		"20260716005855",
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
			t.Fatalf("fresh migration %s did not load", version)
		}
		if strings.Contains(source, "-- +goose Down") {
			t.Fatalf("migration %s contains a destructive down section", version)
		}
	}
}

func TestSystemResourceLayoutDomainMigrationCoversEveryAuthoritySurface(t *testing.T) {
	sources := systemResourceLayoutMigrationSources(t)
	constraintSource := sources["20260716005800"]
	for _, constraint := range []string{
		"domain_cutovers_domain_check",
		"cutover_group_members_domain_check",
		"cutover_transition_events_domain_check",
		"contract_retirements_domain_check",
	} {
		if !strings.Contains(constraintSource, "DROP CONSTRAINT "+constraint) ||
			!strings.Contains(constraintSource, "VALIDATE CONSTRAINT "+constraint) {
			t.Fatalf("constraint migration does not replace and validate %s", constraint)
		}
	}
	if strings.Count(constraintSource, "'system_resource_layout'") != 4 {
		t.Fatal("constraint migration does not expand every persisted domain allowlist")
	}
	for _, marker := range []string{
		"CONSTRAINT domain_cutovers_authority_generation_unique\n" +
			"        UNIQUE (scope_kind, scope_id, domain, generation)",
		"CONSTRAINT cutover_groups_authority_generation_unique\n" +
			"        UNIQUE (scope_kind, scope_id, group_id, generation)",
	} {
		if !strings.Contains(constraintSource, marker) {
			t.Fatalf("constraint migration lacks generation authority key %q", marker)
		}
	}
	for version, source := range sources {
		if version == "20260716005800" {
			continue
		}
		if !strings.Contains(source, "DROP FUNCTION platform.") ||
			!strings.Contains(source, "CREATE FUNCTION platform.") ||
			strings.Count(source, "'system_resource_layout'") != 1 {
			t.Fatalf("routine migration %s does not replace one exact allowlist", version)
		}
		if strings.Contains(source, "CREATE OR REPLACE") ||
			strings.Contains(source, "DROP FUNCTION IF EXISTS") {
			t.Fatalf("routine migration %s uses an inexact replacement form", version)
		}
	}
}

func systemResourceLayoutMigrationSources(t *testing.T) map[string]string {
	t.Helper()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]string{}
	for _, document := range documents {
		if !strings.Contains(document.Path(), "/fresh/202607160058") {
			continue
		}
		name := path.Base(document.Path())
		sources[name[:14]] = string(document.Content())
	}
	if len(sources) != 7 {
		t.Fatalf("forward migration count = %d, want 7", len(sources))
	}
	return sources
}
