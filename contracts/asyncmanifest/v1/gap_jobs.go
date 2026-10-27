package asyncmanifestv1

import (
	"fmt"
	"sort"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

const (
	GapJobPayloadRevision = 1
	GapJobMaximumAttempts = jobsv1.MaximumMaximumAttempts
	gapJobKindPrefix      = "platform_async.event_gap_reconcile."
)

func GapJobPolicy(domain string) (jobsv1.KindPolicy, error) {
	policy, err := jobsv1.NewKindPolicy(
		gapJobKindPrefix+domain, GapJobPayloadRevision, domain,
		jobsv1.EffectDatabaseTransactional, GapJobMaximumAttempts,
	)
	if err != nil {
		return jobsv1.KindPolicy{}, fmt.Errorf("build event gap job policy: %w", err)
	}
	return policy, nil
}

func deriveGapJobPolicies(
	consumers eventsv1.ConsumerCatalog,
) ([]jobsv1.KindPolicy, error) {
	domains := make(map[string]struct{})
	for _, consumer := range consumers.Policies() {
		domains[consumer.OwnerDomain] = struct{}{}
	}
	ordered := make([]string, 0, len(domains))
	for domain := range domains {
		ordered = append(ordered, domain)
	}
	sort.Strings(ordered)
	policies := make([]jobsv1.KindPolicy, 0, len(ordered))
	for _, domain := range ordered {
		policy, err := GapJobPolicy(domain)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}
	return policies, nil
}
