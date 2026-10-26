package postgres

import (
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

type repository struct {
	database    boundDatabase
	authority   authorityStamp
	principalID uuid.UUID
	operationID string
	jobs        jobsv1.Catalog
	producers   eventsv1.ProducerCatalog
}

type consumerIdentity struct{ marker byte }

type consumerRepository struct {
	database        boundDatabase
	authority       authorityStamp
	gapPolicy       jobsv1.KindPolicy
	producers       eventsv1.ProducerCatalog
	consumers       eventsv1.ConsumerCatalog
	consumer        string
	consumerVersion int
	identity        *consumerIdentity
}

func (repository *consumerRepository) valid() bool {
	return repository != nil && repository.identity != nil &&
		repository.identity.marker == 1 && repository.consumer != "" &&
		repository.consumerVersion > 0 && repository.gapPolicy.Validate() == nil
}

func resolveConsumerGapPolicy(
	stamp authorityStamp,
	jobs jobsv1.Catalog,
	producers eventsv1.ProducerCatalog,
	consumers eventsv1.ConsumerCatalog,
	consumer string,
	consumerVersion int,
) (jobsv1.KindPolicy, error) {
	policy, err := consumers.ResolveIdentity(consumer, consumerVersion)
	if err != nil || policy.OwnerDomain != stamp.domain || len(producers.Policies()) == 0 {
		return jobsv1.KindPolicy{}, ErrConsumerBindingInvalid
	}
	gapPolicy, err := platformasync.EventGapJobPolicy(policy.OwnerDomain)
	if err != nil {
		return jobsv1.KindPolicy{}, ErrConsumerBindingInvalid
	}
	registered, err := jobs.Resolve(gapPolicy.Kind, gapPolicy.PayloadRevision)
	if err != nil || registered != gapPolicy {
		return jobsv1.KindPolicy{}, ErrConsumerBindingInvalid
	}
	return gapPolicy, nil
}
