package asyncmanifestv1

import "testing"

func TestAnalyticsPoliciesAreClosedAndValid(t *testing.T) {
	jobs := AnalyticsJobPolicies()
	if len(jobs) != 7 {
		t.Fatalf("analytics job count = %d", len(jobs))
	}
	for _, policy := range jobs {
		if policy.OwnerDomain != "analytics" || policy.Validate() != nil {
			t.Fatalf("analytics job policy = %+v", policy)
		}
	}
	events := AnalyticsProducerPolicies()
	if len(events) != 4 {
		t.Fatalf("analytics event count = %d", len(events))
	}
	for _, policy := range events {
		if policy.OwnerDomain != "analytics" || policy.Validate() != nil {
			t.Fatalf("analytics event policy = %+v", policy)
		}
	}
}
