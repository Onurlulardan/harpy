package billing

import (
	"strings"
	"testing"
)

func TestV1BaseAdoptionMigrationsBindSignedOwnerComposite(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	var combined strings.Builder
	for _, document := range documents {
		if strings.Contains(document.Path(), "/fresh/2026072311300") {
			combined.Write(document.Content())
		}
	}
	for _, token := range []string{
		"platform.billing_v1_base_adoption_apply_v1",
		"platform.billing_v1_base_adoption_apply_and_record_v1",
		"billing.v1_base_adoption.v1",
		"platform.cutover_prepare_batch_v1",
		"platform.cutover_record_batch_v1",
		"platform.billing_v1_base_adoption_assert_v1",
	} {
		if !strings.Contains(combined.String(), token) {
			t.Fatalf("Billing V1 base token %q is missing", token)
		}
	}
}
