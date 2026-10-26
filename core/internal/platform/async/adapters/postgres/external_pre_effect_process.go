package postgres

import (
	"context"
	"errors"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var ErrExternalPreEffectUnavailable = errors.New("external pre-effect store is unavailable")

type ExternalPreEffectStore interface {
	Start(
		context.Context, JobClaim, ExternalEffectRequestSpec,
	) (ExternalEffectPermit, error)
	FailBeforeEffect(
		context.Context, JobClaim, JobFailure,
	) (JobFailureResult, error)
}

type externalPreEffectStore struct{ worker *workerRepository }

func BindExternalPreEffect(store WorkerStore) (ExternalPreEffectStore, error) {
	worker, ok := store.(*workerRepository)
	if !ok || !worker.valid() || worker.coordinator == nil {
		return nil, ErrExternalPreEffectUnavailable
	}
	return &externalPreEffectStore{worker: worker}, nil
}

func (store *externalPreEffectStore) Start(
	ctx context.Context,
	claim JobClaim,
	spec ExternalEffectRequestSpec,
) (ExternalEffectPermit, error) {
	repository := store.worker
	tenantID, tenantScoped := claim.TenantID()
	if ctx == nil || !repository.valid() || repository.coordinator == nil ||
		!claim.valid(repository.identity) || !tenantScoped ||
		claim.EffectClass() != "external" {
		return ExternalEffectPermit{}, ErrJobClaimInvalid
	}
	var permit ExternalEffectPermit
	err := repository.coordinator.InJobTenant(
		ctx, tenantID,
		func(scoped context.Context, scope platformtx.JobTenantScope) error {
			store, err := platformtx.BindJob(scope, repository.databaseJobBinder)
			if err != nil || store == nil {
				return ErrWorkerStoreUnavailable
			}
			if err := store.assertAuthority(scoped, claim); err != nil {
				return err
			}
			permit, err = store.startExternal(
				scoped, repository.identity, claim, spec,
			)
			return err
		},
	)
	return permit, err
}

func (store *externalPreEffectStore) FailBeforeEffect(
	ctx context.Context,
	claim JobClaim,
	failure JobFailure,
) (JobFailureResult, error) {
	repository := store.worker
	tenantID, tenantScoped := claim.TenantID()
	if ctx == nil || !repository.valid() || repository.coordinator == nil ||
		!claim.valid(repository.identity) || !tenantScoped ||
		claim.EffectClass() != "external" {
		return JobFailureResult{}, ErrJobClaimInvalid
	}
	var result JobFailureResult
	err := repository.coordinator.InJobTenant(
		ctx, tenantID,
		func(scoped context.Context, scope platformtx.JobTenantScope) error {
			store, err := platformtx.BindJob(scope, repository.databaseJobBinder)
			if err != nil || store == nil {
				return ErrWorkerStoreUnavailable
			}
			if err := store.assertAuthority(scoped, claim); err != nil {
				return err
			}
			result, err = store.failBeforeExternal(scoped, claim, failure)
			return err
		},
	)
	return result, err
}
