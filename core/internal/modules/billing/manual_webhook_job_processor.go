package billing

import (
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

func newManualOnlyWebhookJobProcessor(
	catalog jobsv1.Catalog,
	store jobexecution.DatabaseJobStore,
) (jobexecution.Processor, error) {
	policy, err := catalog.Resolve(providerapp.WebhookProcessorJobKind, 1)
	if err != nil || policy.OwnerDomain != "billing" ||
		policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
		policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts {
		return nil, fmt.Errorf("manual-only billing webhook job policy is invalid")
	}
	binding, err := newManualOnlyWebhookJobBinding()
	if err != nil {
		return nil, fmt.Errorf("build manual-only billing webhook binding: %w", err)
	}
	processor, err := jobexecution.NewDatabaseProcessor(policy, store, binding)
	if err != nil {
		return nil, fmt.Errorf("build manual-only billing webhook processor: %w", err)
	}
	return processor, nil
}
