package migrator

import "testing"

func TestCutoverApplyRegistryContainsEveryProductionContract(t *testing.T) {
	registry, err := newCutoverApplyRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, contractID := range []string{
		"platform.cutover_probe.v1",
		"platform.v1_environment_plan.v1",
		"tenancy.v1_environment_adoption.v1",
		"iam.v1_protected_authority_adoption.v1",
		"billing.v1_base_adoption.v1",
		"platform.v1_environment_completion.v1",
		"billing.v1_adoption.global.v1",
		"iam.v1_adoption.v1",
		"billing.v1_adoption.tenant.v1",
		"analytics.v1_adoption.v1",
		"catalog.v1_adoption.v1",
		"work.v1_adoption.v1",
		"documents.v1_adoption.v1",
		"notifications.v1_adoption.v1",
	} {
		query, found := registry.Lookup(contractID)
		if !found || query == "" {
			t.Fatalf("cutover apply contract %q is missing", contractID)
		}
	}
	if _, found := registry.Lookup("unregistered.owner.v1"); found {
		t.Fatal("cutover apply registry accepted an unregistered owner contract")
	}
}
