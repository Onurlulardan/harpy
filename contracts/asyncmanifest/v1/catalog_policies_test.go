package asyncmanifestv1

import "testing"

func TestCatalogPoliciesAreClosedAndValid(t *testing.T) {
	jobs := CatalogJobPolicies()
	if len(jobs) != 1 || jobs[0].OwnerDomain != "catalog" || jobs[0].Validate() != nil {
		t.Fatalf("catalog jobs = %+v", jobs)
	}
	events := CatalogProducerPolicies()
	if len(events) != 5 {
		t.Fatalf("catalog event count = %d", len(events))
	}
	for _, policy := range events {
		if policy.OwnerDomain != "catalog" || policy.Validate() != nil {
			t.Fatalf("catalog event policy = %+v", policy)
		}
	}
}
