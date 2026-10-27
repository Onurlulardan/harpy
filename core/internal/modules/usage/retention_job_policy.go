package usage

import (
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application"
)

func RetentionJobPolicy() jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(application.RetentionJobKind,
		application.RetentionPayloadRevision, application.RetentionAuthorityDomain,
		jobsv1.EffectDatabaseTransactional, jobsv1.DefaultMaximumAttempts)
	return policy
}

func RetentionJobPolicies() []jobsv1.KindPolicy {
	return []jobsv1.KindPolicy{RetentionJobPolicy()}
}

func ValidateRetentionJobPolicy(policy jobsv1.KindPolicy) error {
	if policy.Validate() != nil || policy != RetentionJobPolicy() {
		return fmt.Errorf("usage retention job policy is invalid")
	}
	return nil
}
