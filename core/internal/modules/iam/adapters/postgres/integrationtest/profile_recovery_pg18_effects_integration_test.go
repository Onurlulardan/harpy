//go:build integration

package integrationtest

import (
	"context"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type recoveryPGEffects struct {
	catalogs platformasync.WriteCatalogs
}

func (effects recoveryPGEffects) Enqueue(
	ctx context.Context, scope platformtx.TenantScope,
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
		JobID: first.JobID.String(), Kind: "iam.profile.backfill",
		PayloadRevision: 1, TenantID: &tenantID,
		Authority: jobsv1.Authority{
			ScopeKind: jobsv1.ScopeTenant, ScopeID: tenantID,
			Domain: authority.Domain(), CutoverGeneration: authority.Generation(),
			WriterOrigin: jobsv1.WriterV2, GroupID: authority.GroupID(),
			GroupGeneration: authority.GroupGeneration(),
		},
		PayloadReference: first.PayloadReference,
		PayloadSHA256:    first.PayloadSHA256,
		RunAt:            first.RunAt.UTC().Format(time.RFC3339Nano),
		DedupeKey:        first.DedupeKey, MaximumAttempts: jobsv1.DefaultMaximumAttempts,
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
