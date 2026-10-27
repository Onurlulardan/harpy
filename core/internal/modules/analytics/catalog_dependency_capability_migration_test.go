package analytics

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func TestCatalogDependencyCapabilityAllowsCanonicalEmptyValue(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	for _, origin := range []string{"fresh", "v1_adopted"} {
		helper := analyticsMigrationContent(
			t, documents, origin,
			"20260730143001_create_analytics_catalog_projection_helpers_v2",
		)
		materializer := analyticsMigrationContent(
			t, documents, origin,
			"20260730143002_create_analytics_catalog_endpoint_materializer_v2",
		)
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

func analyticsMigrationContent(
	t *testing.T,
	documents []migration.Document,
	origin string,
	logical string,
) string {
	t.Helper()
	needle := "/tenant/" + origin + "/" + logical + "." + origin + ".sql"
	for _, document := range documents {
		path := strings.ReplaceAll(document.Path(), `\`, "/")
		if strings.HasSuffix(path, needle) {
			return string(document.Content())
		}
	}
	t.Fatalf("Analytics migration is missing: %s", needle)
	return ""
}
