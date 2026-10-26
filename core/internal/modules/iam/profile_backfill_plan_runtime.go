package iam

import (
	"context"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	profilebackfillpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/profilebackfill"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const ProfileBackfillJobKind = "iam.profile.backfill"

type ProfileBackfillPlanRuntime struct {
	service *profilebackfillapp.Service
}

type profileBackfillEffects struct {
	catalogs platformasync.WriteCatalogs
}

func NewProfileBackfillPlanRuntime(
	catalogs platformasync.WriteCatalogs,
) (*ProfileBackfillPlanRuntime, error) {
	if _, _, err := catalogs.Snapshot(); err != nil {
		return nil, profilebackfillapp.ErrEffects
	}
	store, err := profilebackfillpostgres.NewStore()
	if err != nil {
		return nil, err
	}
	service, err := profilebackfillapp.NewService(
		store, profileBackfillEffects{catalogs: catalogs},
	)
	if err != nil {
		return nil, err
	}
	return &ProfileBackfillPlanRuntime{service: service}, nil
}

func (runtime *ProfileBackfillPlanRuntime) Create(
	ctx context.Context,
	scope platformtx.TenantScope,
	command profilebackfillapp.CreateCommand,
) (profilebackfillapp.Result, error) {
	if runtime == nil || runtime.service == nil {
		return profilebackfillapp.Result{}, profilebackfillapp.ErrUnavailable
	}
	return runtime.service.Create(ctx, scope, command)
}

func (effects profileBackfillEffects) Enqueue(
	ctx context.Context,
	scope platformtx.TenantScope,
	first profilebackfillapp.FirstJob,
) error {
	authority, err := platformtx.MintOwnerAuthority(scope, "iam")
	if err != nil || !authority.OwnsScope(scope) {
		return profilebackfillapp.ErrEffects
	}
	jobs, _, err := effects.catalogs.Snapshot()
	if err != nil {
		return profilebackfillapp.ErrEffects
	}
	tenantID := scope.TenantID().String()
	job, err := jobsv1.NewEnqueue(jobsv1.Enqueue{
		JobID: first.JobID.String(), Kind: ProfileBackfillJobKind,
		PayloadRevision: 1, TenantID: &tenantID,
		Authority: jobsv1.Authority{
			ScopeKind: jobsv1.ScopeTenant, ScopeID: tenantID,
			Domain: authority.Domain(), CutoverGeneration: authority.Generation(),
			WriterOrigin: jobsv1.WriterV2, GroupID: authority.GroupID(),
			GroupGeneration: authority.GroupGeneration(),
		},
		PayloadReference: first.PayloadReference, PayloadSHA256: first.PayloadSHA256,
		RunAt: first.RunAt.UTC().Format(time.RFC3339Nano), Priority: 0,
		DedupeKey: first.DedupeKey, MaximumAttempts: jobsv1.DefaultMaximumAttempts,
	}, jobs)
	if err != nil {
		return profilebackfillapp.ErrEffects
	}
	writer, err := asyncpostgres.Bind(scope, authority, effects.catalogs)
	if err != nil {
		return profilebackfillapp.ErrEffects
	}
	result, err := writer.EnqueueJob(ctx, job)
	if err != nil || result.ID != first.JobID.String() {
		return profilebackfillapp.ErrEffects
	}
	return nil
}
