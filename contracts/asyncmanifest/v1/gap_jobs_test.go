package asyncmanifestv1

import (
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

func TestGapJobsAreDerivedOncePerConsumerOwnerDomain(t *testing.T) {
	first, _ := eventsv1.NewConsumerPolicy(
		"workflow.projector", 1, "workflow", "forms.created", 1,
		"workflow.projector.v1", eventsv1.EffectDatabaseTransactional, 8,
	)
	second, _ := eventsv1.NewConsumerPolicy(
		"workflow.audit", 1, "workflow", "forms.updated", 1,
		"workflow.audit.v1", eventsv1.EffectDatabaseTransactional, 8,
	)
	third, _ := eventsv1.NewConsumerPolicy(
		"analytics.projector", 1, "analytics", "forms.created", 1,
		"analytics.projector.v1", eventsv1.EffectDatabaseTransactional, 8,
	)
	consumers, err := eventsv1.NewConsumerCatalog([]eventsv1.ConsumerPolicy{
		first, second, third,
	})
	if err != nil {
		t.Fatal(err)
	}
	policies, err := deriveGapJobPolicies(consumers)
	if err != nil {
		t.Fatal(err)
	}
	if len(policies) != 2 || policies[0].OwnerDomain != "analytics" ||
		policies[1].OwnerDomain != "workflow" {
		t.Fatalf("derived gap policies = %#v", policies)
	}
	for _, policy := range policies {
		if policy.EffectClass != "database_transactional" ||
			policy.MaximumAttempts != GapJobMaximumAttempts {
			t.Fatalf("unsafe gap policy = %#v", policy)
		}
	}
}
