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
	versions := map[string]string{}
	for _, document := range documents {
		documentPath := document.Path()
		if !strings.Contains(documentPath, "/fresh/") {
			continue
		}
		for _, version := range []string{"20260723114004", "20260723114005", "20260723114006"} {
			if strings.Contains(documentPath, version) {
				versions[version] = string(document.Content())
			}
		}
	}
	for _, version := range []string{"20260723114004", "20260723114005", "20260723114006"} {
		if versions[version] == "" {
			t.Fatalf("fresh migration %s did not load", version)
		}
	}
	combined := versions["20260723114004"] +
		versions["20260723114005"] + versions["20260723114006"]
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
