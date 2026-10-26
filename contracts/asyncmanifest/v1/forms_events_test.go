package asyncmanifestv1

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
)

func TestFormsProducerPoliciesPinSchemaBytes(t *testing.T) {
	policies := FormsProducerPolicies()
	fixtures := []struct{ eventType, schemaSHA256, file string }{
		{FormsPublicationCompletedEventType, FormsPublicationCompletedSchemaSHA256,
			"forms-publication-completed-v1.schema.json"},
		{FormsRecordCreatedEventType, FormsRecordCreatedSchemaSHA256,
			"forms-record-created-v1.schema.json"},
		{FormsRecordDeletedEventType, FormsRecordDeletedSchemaSHA256,
			"forms-record-deleted-v1.schema.json"},
		{FormsRecordUpdatedEventType, FormsRecordUpdatedSchemaSHA256,
			"forms-record-updated-v1.schema.json"},
	}
	if len(policies) != len(fixtures) {
		t.Fatalf("Forms producer policy count = %d", len(policies))
	}
	for index, fixture := range fixtures {
		policy := policies[index]
		if policy.EventType != fixture.eventType || policy.OwnerDomain != "forms" ||
			policy.EventVersion != 1 {
			t.Fatalf("Forms producer policy %d = %#v", index, policy)
		}
		document, err := os.ReadFile("../../events/v1/" + fixture.file)
		if err != nil {
			t.Fatal(err)
		}
		digest := fmt.Sprintf("%x", sha256.Sum256(document))
		if digest != fixture.schemaSHA256 || policy.PayloadSchemaSHA256 != digest {
			t.Fatalf("Forms schema %s digest = %s", fixture.file, digest)
		}
	}
}
