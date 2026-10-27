package tenancy

import (
	"strings"
	"testing"
)

func TestV1EnvironmentAdoptionMigrationsBindSignedOwnerComposite(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	var combined strings.Builder
	for _, document := range documents {
		if strings.Contains(document.Path(), "/fresh/2026072311200") {
			combined.Write(document.Content())
		}
	}
	for _, token := range []string{
		"platform.tenancy_v1_environment_apply_v1",
		"platform.tenancy_v1_environment_apply_and_record_v1",
		"tenancy.v1_environment_adoption.v1",
		"platform.cutover_prepare_batch_v1",
		"platform.cutover_record_batch_v1",
		"platform.tenancy_v1_adopt_environment_v1",
	} {
		if !strings.Contains(combined.String(), token) {
			t.Fatalf("tenancy V1 adoption token %q is missing", token)
		}
	}
}
