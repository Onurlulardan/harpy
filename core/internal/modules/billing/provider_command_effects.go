package billing

import (
	"context"
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	providerpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/provider"
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type providerCommandEffects struct {
	tenantID  uuid.UUID
	authority platformtx.OwnerAuthority
	jobs      jobsv1.Catalog
	writer    asyncpostgres.Writer
}

func newProviderCommandEffects(
	scope platformtx.TenantScope,
	authority platformtx.OwnerAuthority,
	catalogs platformasync.WriteCatalogs,
) (*providerCommandEffects, error) {
	jobs, _, err := catalogs.Snapshot()
	if err != nil || !authority.OwnsScope(scope) ||
		authority.Domain() != "billing" {
		return nil, providerapp.ErrConfiguration
	}
	writer, err := asyncpostgres.Bind(scope, authority, catalogs)
	if err != nil {
		return nil, fmt.Errorf("bind billing provider command async writer: %w", err)
	}
	return &providerCommandEffects{
		tenantID: scope.TenantID(), authority: authority,
		jobs: jobs, writer: writer,
	}, nil
}

func (effects *providerCommandEffects) EnqueueCommandRoute(
	ctx context.Context,
	snapshot providerapp.CommandSnapshot,
) error {
	if effects == nil || snapshot.TenantID != effects.tenantID {
		return providerapp.ErrCommandState
	}
	command, err := providerapp.NewProviderCommandJob(
		snapshot.Provider, snapshot.CommandID,
		fmt.Sprintf("%x", snapshot.RequestSHA256), snapshot.RequestedAt,
	)
	if err != nil {
		return err
	}
	finalizer, err := providerapp.NewProviderFinalizeJob(
		snapshot.CommandID, fmt.Sprintf("%x", snapshot.RequestSHA256),
		snapshot.RequestedAt,
	)
	if err != nil {
		return err
	}
	if err := effects.enqueue(ctx, snapshot.CommandID, command); err != nil {
		return err
	}
	return effects.enqueue(ctx, snapshot.FinalizerJobID, finalizer)
}

func (effects *providerCommandEffects) enqueue(
	ctx context.Context,
	jobID uuid.UUID,
	job providerapp.AsyncJob,
) error {
	policy := job.Policy()
	tenantID := effects.tenantID.String()
	var effectKey *string
	if job.EffectKey() != "" {
		value := job.EffectKey()
		effectKey = &value
	}
	value, err := jobsv1.NewEnqueue(jobsv1.Enqueue{
		JobID: jobID.String(), Kind: policy.Kind,
		PayloadRevision: policy.PayloadRevision, TenantID: &tenantID,
		Authority:        subscriptionJobAuthority(effects.authority),
		PayloadReference: job.PayloadReference(),
		PayloadSHA256:    job.PayloadSHA256(),
		RunAt:            job.RunAt().Format(time.RFC3339Nano), Priority: 100,
		DedupeKey: job.DedupeKey(), EffectIdempotencyKey: effectKey,
		MaximumAttempts: policy.MaximumAttempts,
	}, effects.jobs)
	if err != nil {
		return err
	}
	result, err := effects.writer.EnqueueJob(ctx, value)
	if err != nil || result.ID != jobID.String() {
		return fmt.Errorf("enqueue billing provider job %s: %w", jobID, err)
	}
	return nil
}

var _ providerpostgres.CommandEffects = (*providerCommandEffects)(nil)
