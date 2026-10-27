package asyncmanifestv1

import "testing"

func TestWorkProducerPolicyPinsSchemaBytes(t *testing.T) {
	reminderDigest := schemaDigest(t, "../../events/v1/work-reminder-due-v1.schema.json")
	mutationDigest := schemaDigest(t, "../../events/v1/work-mutation-v1.schema.json")
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
