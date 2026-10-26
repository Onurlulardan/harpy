package notifications

import (
	"context"

	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

type deliveryPreEffect struct {
	store asyncpostgres.ExternalPreEffectStore
}

func (preEffect deliveryPreEffect) Start(ctx context.Context,
	claim asyncpostgres.JobClaim, spec asyncpostgres.ExternalEffectRequestSpec,
) (jobexecution.StartedEffect, error) {
	permit, err := preEffect.store.Start(ctx, claim, spec)
	if err != nil {
		return jobexecution.StartedEffect{}, err
	}
	return jobexecution.NewStartedEffect(permit)
}

func (preEffect deliveryPreEffect) FailBeforeEffect(ctx context.Context,
	claim asyncpostgres.JobClaim, failure asyncpostgres.JobFailure,
) (asyncpostgres.JobFailureResult, error) {
	return preEffect.store.FailBeforeEffect(ctx, claim, failure)
}

var _ jobexecution.PreEffect = deliveryPreEffect{}
