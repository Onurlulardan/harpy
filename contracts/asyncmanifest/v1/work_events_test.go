package asyncmanifestv1

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
)

func TestWorkProducerPolicyPinsSchemaBytes(t *testing.T) {
	reminder, err := os.ReadFile("../../events/v1/work-reminder-due-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	reminderDigest := fmt.Sprintf("%x", sha256.Sum256(reminder))
	mutation, err := os.ReadFile("../../events/v1/work-mutation-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	mutationDigest := fmt.Sprintf("%x", sha256.Sum256(mutation))
	policies := WorkProducerPolicies()
	if len(policies) != 1+len(workMutationEventTypes()) ||
		policies[0].EventType != WorkReminderDueEventType ||
		policies[0].EventVersion != 1 || policies[0].OwnerDomain != "work" ||
		reminderDigest != WorkReminderDueSchemaSHA256 ||
		policies[0].PayloadSchemaSHA256 != reminderDigest {
		t.Fatalf("Work reminder policy does not pin schema: %#v", policies)
	}
	if mutationDigest != WorkMutationSchemaSHA256 {
		t.Fatalf("Work mutation schema digest = %s", mutationDigest)
	}
	for index, eventType := range workMutationEventTypes() {
		policy := policies[index+1]
		if policy.EventType != eventType || policy.EventVersion != 1 ||
			policy.OwnerDomain != "work" || policy.PayloadSchemaSHA256 != mutationDigest {
			t.Fatalf("Work mutation policy %d = %#v", index, policy)
		}
	}
}
