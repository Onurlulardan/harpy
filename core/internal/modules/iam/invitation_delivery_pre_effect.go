package iam

import (
	"context"

	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

type invitationDeliveryPreEffect struct {
	store asyncpostgres.ExternalPreEffectStore
}

func (preEffect invitationDeliveryPreEffect) Start(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
	spec asyncpostgres.ExternalEffectRequestSpec,
) (jobexecution.StartedEffect, error) {
	permit, err := preEffect.store.Start(ctx, claim, spec)
	if err != nil {
		return jobexecution.StartedEffect{}, err
	}
	return jobexecution.NewStartedEffect(permit)
}

func (preEffect invitationDeliveryPreEffect) FailBeforeEffect(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
	failure asyncpostgres.JobFailure,
) (asyncpostgres.JobFailureResult, error) {
	return preEffect.store.FailBeforeEffect(ctx, claim, failure)
}

var _ jobexecution.PreEffect = invitationDeliveryPreEffect{}
