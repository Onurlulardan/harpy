package migration

import (
	"strings"
	"testing"
)

func TestLoadStartupRegistryBuildsFreshManifest(t *testing.T) {
	first := freshTestDocument(t, testFreshPath, 1)
	second := freshTestDocument(
		t, "core/internal/modules/audit/migrations/tenant/fresh/"+
			"20260801040000_create_audit_change_log.fresh.sql", 2,
	)
	replaced := freshTestDocument(
		t, "core/internal/modules/iam/migrations/tenant/fresh/"+
			"20260729100025_create_iam_installation_tenant_capability_entrypoints_v1.fresh.sql", 3,
	)
	terminal := freshTestDocument(
		t, "core/internal/modules/iam/migrations/tenant/supersessions/fresh/"+
			"20260729100025_create_iam_installation_tenant_capability_entrypoints_v1.fresh.sql", 4,
	)
	registry, err := LoadStartupRegistry([]Document{second, replaced, terminal, first})
	if err != nil {
		t.Fatal(err)
	}
	paths := registry.FreshPaths()
	if registry.FreshEntryCount() != 3 || len(paths) != 3 || len(registry.Digests()) != 1 {
		t.Fatalf("fresh registry count = %d paths = %v", registry.FreshEntryCount(), paths)
	}
	for _, documentPath := range paths {
		if strings.Contains(documentPath, "v1_adopted") {
			t.Fatalf("adopted path entered fresh registry: %s", documentPath)
		}
	}
	if !strings.Contains(paths[1], "/supersessions/fresh/") {
		t.Fatalf("terminal IAM path = %s", paths[1])
	}
	if err = registry.AuthorizeDigest(registry.Digests()[0]); err != nil {
		t.Fatal(err)
	}
	if _, err = LoadStartupRegistry(nil); err == nil ||
		!strings.Contains(err.Error(), "Harpy fresh schema is not present") {
		t.Fatalf("empty registry error = %v", err)
	}
}
