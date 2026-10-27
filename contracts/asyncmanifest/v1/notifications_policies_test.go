package asyncmanifestv1

import (
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestNotificationsJobPoliciesAreClosedEffects(t *testing.T) {
	policies := NotificationsJobPolicies()
	if len(policies) != 3 {
		t.Fatalf("policy count = %d", len(policies))
	}
	want := []struct {
		kind   string
		effect jobsv1.EffectClass
	}{
		{"notifications.email.delivery", jobsv1.EffectExternal},
		{NotificationsRetentionJobKind, jobsv1.EffectDatabaseTransactional},
		{"notifications.telegram.delivery", jobsv1.EffectExternal},
	}
	for index, policy := range policies {
		if err := policy.Validate(); err != nil {
			t.Fatal(err)
		}
		if policy.Kind != want[index].kind || policy.OwnerDomain != "notifications" ||
			policy.EffectClass != want[index].effect || policy.MaximumAttempts != 8 {
			t.Fatalf("policy = %+v", policy)
		}
	}
}
