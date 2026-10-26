package cutover

import (
	"strings"
	"testing"
)

func TestV1EnvironmentAdoptionMigrationsBindSignedCompositeContracts(t *testing.T) {
	t.Parallel()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	versions := map[string]map[string]string{}
	for _, document := range documents {
		path := document.Path()
		for _, version := range []string{"20260723114004", "20260723114005", "20260723114006"} {
			if !strings.Contains(path, version) {
				continue
			}
			origin := "fresh"
			if strings.Contains(path, "/v1_adopted/") {
				origin = "v1_adopted"
			}
			if versions[version] == nil {
				versions[version] = map[string]string{}
			}
			versions[version][origin] = string(document.Content())
		}
	}
	for version, origins := range versions {
		if len(origins) != 2 || origins["fresh"] != origins["v1_adopted"] {
			t.Fatalf("migration %s is not an exact origin pair", version)
		}
	}
	combined := versions["20260723114004"]["fresh"] +
		versions["20260723114005"]["fresh"] + versions["20260723114006"]["fresh"]
	for _, token := range []string{
		"platform-v1-environment-adoption.v1",
		"platform-v1-environment-completion.v1",
		"platform.cutover_prepare_batch_v1",
		"platform.cutover_record_batch_v1",
		"platform.platform_v1_environment_plan_apply_v1",
		"platform.platform_v1_environment_completion_apply_v1",
	} {
		if !strings.Contains(combined, token) {
			t.Fatalf("V1 environment migration token %q is missing", token)
		}
	}
}
