package documents

import (
	"context"

	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

type adoptionCopyPreEffect struct {
	store asyncpostgres.ExternalPreEffectStore
}

func (preEffect adoptionCopyPreEffect) Start(ctx context.Context,
	claim asyncpostgres.JobClaim, spec asyncpostgres.ExternalEffectRequestSpec,
) (jobexecution.StartedEffect, error) {
	permit, err := preEffect.store.Start(ctx, claim, spec)
	if err != nil {
		return jobexecution.StartedEffect{}, err
	}
	return jobexecution.NewStartedEffect(permit)
}

func (preEffect adoptionCopyPreEffect) FailBeforeEffect(ctx context.Context,
	claim asyncpostgres.JobClaim, failure asyncpostgres.JobFailure,
) (asyncpostgres.JobFailureResult, error) {
	return preEffect.store.FailBeforeEffect(ctx, claim, failure)
}

var _ jobexecution.PreEffect = adoptionCopyPreEffect{}
