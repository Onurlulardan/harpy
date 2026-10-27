package catalogowner

import (
	"strings"
	"testing"
)

func TestCatalogOwnerUsesVersionScopedProjectionAndActivation(t *testing.T) {
	for label, source := range map[string]string{
		"provision": provisionSQL,
		"read":      readSQL,
	} {
		if strings.Contains(source, "iam_catalog_project_v1(") {
			t.Fatalf("%s SQL retained component-scoped IAM catalog routine", label)
		}
	}
	if !strings.Contains(provisionSQL, "iam_catalog_project_v2(") {
		t.Fatal("provision SQL lacks IAM version projection")
	}
	if !strings.Contains(applySQL, "iam_catalog_component_apply_v1(") ||
		strings.Contains(applySQL, "iam_catalog_component_apply_v2(") {
		t.Fatal("apply SQL lacks the routed IAM lifecycle entrypoint")
	}
	for _, marker := range []string{
		"iam_catalog_component_candidates",
		"COALESCE(candidate.content_sha256,component.content_sha256)",
	} {
		if !strings.Contains(readSQL, marker) {
			t.Fatalf("read SQL lacks %q", marker)
		}
	}
}
