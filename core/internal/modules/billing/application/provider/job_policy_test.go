package providerapp

import (
	"testing"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestJobPoliciesMatchADR0011AndHaveUniqueIdentities(t *testing.T) {
	policies := JobPolicies()
	catalog, err := jobsv1.NewCatalog(policies)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Policies()) != 4 {
		t.Fatalf("policy count=%d", len(catalog.Policies()))
	}
	for _, policy := range policies {
		if policy.OwnerDomain != "billing" || policy.PayloadRevision != 1 ||
			policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts || policy.Validate() != nil {
			t.Fatalf("invalid policy: %+v", policy)
		}
	}
	if WebhookProcessorPolicy().EffectClass != jobsv1.EffectDatabaseTransactional ||
		ProviderCommandPolicy().EffectClass != jobsv1.EffectExternal ||
		ProviderFinalizePolicy().EffectClass != jobsv1.EffectDatabaseTransactional ||
		ProviderReconciliationPolicy().EffectClass != jobsv1.EffectExternal {
		t.Fatal("billing provider jobs have incorrect effect classes")
	}
}

func TestJobPoliciesAreCanonicalReleasePolicies(t *testing.T) {
	manifest, err := asyncmanifestv1.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, policy := range JobPolicies() {
		resolved, err := manifest.Jobs().Resolve(policy.Kind, policy.PayloadRevision)
		if err != nil || resolved != policy {
			t.Fatalf("provider job policy %s drifted: %#v/%v", policy.Kind, resolved, err)
		}
	}
}
