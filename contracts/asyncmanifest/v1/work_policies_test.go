package asyncmanifestv1

import (
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestWorkJobPoliciesAreClosedAndTransactional(t *testing.T) {
	policies := WorkJobPolicies()
	if len(policies) != 4 {
		t.Fatalf("Work policy count = %d", len(policies))
	}
	for _, policy := range policies {
		if policy.Validate() != nil || policy.OwnerDomain != "work" ||
			policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
			policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts {
			t.Fatalf("invalid Work policy: %#v", policy)
		}
	}
}
