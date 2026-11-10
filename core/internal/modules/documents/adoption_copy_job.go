package documents

import (
	"context"
	"errors"
	"fmt"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	"github.com/google/uuid"
)

var adoptionCopyEffectNamespace = uuid.MustParse("6b1e2a44-7c0d-4e8a-9f31-2d5c8a7b0e16")

var ErrAdoptionCopyUnavailable = errors.New("documents adoption copy effect is unavailable")

type AdoptionCopyEffect interface {
	Execute(context.Context, foundation.ID, foundation.ID, foundation.ID, foundation.ID) error
}

type adoptionCopyEffect struct{}

func (adoptionCopyEffect) Execute(context.Context, foundation.ID, foundation.ID,
	foundation.ID, foundation.ID,
) error {
	return ErrAdoptionCopyUnavailable
}

func NewAdoptionCopyJobProcessor(release platformasync.ReleaseCatalogs,
	store asyncpostgres.WorkerStore, now func() time.Time,
) (jobexecution.Processor, error) {
	if store == nil || now == nil {
		return nil, ErrDocumentsRuntime
	}
	jobs, _, err := release.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("read Documents adoption copy catalog: %w", err)
	}
	policy, err := jobs.Resolve(asyncmanifestv1.DocumentsAdoptionCopyJobKind, 1)
	if err != nil || policy.OwnerDomain != "documents" ||
		policy.EffectClass != jobsv1.EffectExternal ||
		policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts {
		return nil, ErrDocumentsRuntime
	}
	provider, err := newAdoptionCopyProvider(adoptionCopyEffect{}, now)
	if err != nil {
		return nil, err
	}
	registry, err := effectexecution.NewRegistry([]effectexecution.Provider{provider})
	if err != nil {
		return nil, err
	}
	preEffect, err := asyncpostgres.BindExternalPreEffect(store)
	if err != nil {
		return nil, err
	}
	processor, err := jobexecution.NewExternalProcessor(policy, adoptionCopyPlanner{},
		adoptionCopyPreEffect{store: preEffect}, store, registry)
	if err != nil {
		return nil, fmt.Errorf("build Documents adoption copy processor: %w", err)
	}
	return processor, nil
}

type adoptionCopyPlanner struct{}

func (adoptionCopyPlanner) Plan(_ context.Context,
	claim asyncpostgres.JobClaim,
) (asyncpostgres.ExternalEffectRequestSpec, error) {
	parsed, err := jobcontract.ParseAdoptionCopy(claim.PayloadReference())
	tenantID, tenantScoped := claim.TenantID()
	_, digest, digestErr := jobcontract.AdoptionCopy(
		parsed.TenantID, parsed.RunID, parsed.SourceVersionID)
	effectKey, hasEffectKey := claim.EffectIdempotencyKey()
	wantKey := jobcontract.AdoptionCopyEffectKey(parsed.RunID, parsed.SourceVersionID)
	if err != nil || digestErr != nil || !tenantScoped ||
		parsed.TenantID.String() != tenantID.String() ||
		claim.Kind() != asyncmanifestv1.DocumentsAdoptionCopyJobKind ||
		claim.PayloadRevision() != 1 || claim.PayloadSHA256() != digest ||
		!hasEffectKey || effectKey != wantKey {
		return asyncpostgres.ExternalEffectRequestSpec{},
			fmt.Errorf("Documents adoption copy claim is invalid")
	}
	attemptID := claim.AttemptID()
	effectID := uuid.NewSHA1(adoptionCopyEffectNamespace, attemptID[:])
	return asyncpostgres.NewExternalEffectRequestSpec(effectID,
		"documents.adoption", "copy", claim.JobID().String(), claim.JobID())
}

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
