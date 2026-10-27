package iam

import (
	"strings"
	"testing"
)

func TestV1ProtectedAuthorityMigrationsBindSignedOwnerComposite(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	var combined strings.Builder
	for _, document := range documents {
		if strings.Contains(document.Path(), "/fresh/2026072311100") {
			combined.Write(document.Content())
		}
	}
	for _, token := range []string{
		"platform.iam_v1_protected_authority_apply_v1",
		"platform.iam_v1_protected_authority_apply_and_record_v1",
		"iam.v1_protected_authority_adoption.v1",
		"platform.cutover_prepare_batch_v1",
		"platform.cutover_record_batch_v1",
		"platform.iam_v1_adopt_protected_authority_v1",
	} {
		if !strings.Contains(combined.String(), token) {
			t.Fatalf("IAM V1 authority token %q is missing", token)
		}
	}
}
