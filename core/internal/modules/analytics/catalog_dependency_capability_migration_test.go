package analytics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogDependencyCapabilityAllowsCanonicalEmptyValue(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		helper := analyticsTenantMigration(t, origin,
			"20260730143001_create_analytics_catalog_projection_helpers_v2")
		materializer := analyticsTenantMigration(t, origin,
			"20260730143002_create_analytics_catalog_endpoint_materializer_v2")
		if strings.Count(helper, "NOT BETWEEN 0 AND 128") != 2 ||
			strings.Contains(helper, "NOT BETWEEN 1 AND 128") ||
			!strings.Contains(materializer, "BETWEEN 0 AND 128 AND") ||
			strings.Contains(materializer, "BETWEEN 1 AND 128 AND") {
			t.Fatalf(
				"%s Analytics Catalog revision capability contract differs",
				origin,
			)
		}
	}
}

func analyticsTenantMigration(t *testing.T, origin, logical string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(
		"migrations", "tenant", origin, logical+"."+origin+".sql",
	))
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}
