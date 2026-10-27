package asyncmanifestv1

import (
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestBillingJobPoliciesAreExactReleaseEntries(t *testing.T) {
	manifest, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	expected := BillingJobPolicies()
	for _, policy := range expected {
		resolved, err := manifest.Jobs().Resolve(policy.Kind, policy.PayloadRevision)
		if err != nil || resolved != policy {
			t.Fatalf("billing job policy %s drifted: %#v/%v", policy.Kind, resolved, err)
		}
	}
	if len(expected) != 6 {
		t.Fatalf("billing job policy count = %d", len(expected))
	}
	if expected[0].EffectClass != jobsv1.EffectDatabaseTransactional ||
		expected[1].EffectClass != jobsv1.EffectDatabaseTransactional ||
		expected[2].EffectClass != jobsv1.EffectExternal ||
		expected[3].EffectClass != jobsv1.EffectDatabaseTransactional ||
		expected[4].EffectClass != jobsv1.EffectExternal ||
		expected[5].EffectClass != jobsv1.EffectDatabaseTransactional {
		t.Fatalf("billing job effect classes = %#v", expected)
	}
}

func TestBillingProducerPoliciesMatchVersionedSchemas(t *testing.T) {
	manifest, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"billing.entitlements.changed":    "billing-entitlements-change-v1.schema.json",
		"billing.invoice.finalized":       "billing-invoice-finalized-v1.schema.json",
		"billing.payment.changed":         "billing-payment-change-v1.schema.json",
		"billing.subscription.changed":    "billing-subscription-change-v1.schema.json",
		"billing.usage.threshold_reached": "billing-usage-threshold-reached-v1.schema.json",
	}
	for eventType, file := range files {
		policy, err := manifest.Producers().Resolve(eventType, 1)
		if err != nil {
			t.Fatalf("resolve %s: %v", eventType, err)
		}
		digest := schemaDigest(t, "../../events/v1/"+file)
		if policy.OwnerDomain != "billing" || policy.PayloadSchemaSHA256 != digest {
			t.Fatalf("billing producer %s = %#v", eventType, policy)
		}
	}
	if len(BillingProducerPolicies()) != len(files) {
		t.Fatalf("billing producer count = %d", len(BillingProducerPolicies()))
	}
}
