package postgres

import (
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func newDatabaseJobChildEnqueue(
	claim JobClaim,
	child platformasync.DatabaseJobChildInput,
	catalog jobsv1.Catalog,
) (jobsv1.Enqueue, jobsv1.KindPolicy, error) {
	tenantID, tenantScoped := claim.TenantID()
	if !claim.valid(claim.authority.owner) || !tenantScoped ||
		claim.authority.scopeKind != "tenant" || claim.authority.scopeID != tenantID ||
		!child.Valid() || child.JobID == claim.JobID() || child.Kind == claim.Kind() ||
		child.RunAt.Before(claim.StartedAt()) {
		return jobsv1.Enqueue{}, jobsv1.KindPolicy{},
			platformasync.ErrDatabaseJobOutcomeInvalid
	}
	policy, err := catalog.Resolve(child.Kind, child.PayloadRevision)
	if err != nil || policy.OwnerDomain != claim.authority.domain {
		return jobsv1.Enqueue{}, jobsv1.KindPolicy{},
			platformasync.ErrDatabaseJobOutcomeInvalid
	}
	var effectKey *string
	if child.EffectIdempotencyKey != "" {
		key := child.EffectIdempotencyKey
		effectKey = &key
	}
	tenant := tenantID.String()
	value, err := jobsv1.NewEnqueue(jobsv1.Enqueue{
		JobID: child.JobID.String(), Kind: child.Kind,
		PayloadRevision: child.PayloadRevision, TenantID: &tenant,
		Authority: jobsv1.Authority{
			ScopeKind: jobsv1.ScopeTenant, ScopeID: tenant,
			Domain:            claim.authority.domain,
			CutoverGeneration: claim.authority.generation,
			WriterOrigin:      jobsv1.WriterV2,
			GroupID:           claim.authority.groupID,
			GroupGeneration:   claim.authority.groupGeneration,
		},
		PayloadReference: child.PayloadReference,
		PayloadSHA256:    child.PayloadSHA256,
		RunAt:            child.RunAt.Format(time.RFC3339Nano), Priority: child.Priority,
		DedupeKey: child.DedupeKey, EffectIdempotencyKey: effectKey,
		MaximumAttempts: policy.MaximumAttempts,
	}, catalog)
	if err != nil {
		return jobsv1.Enqueue{}, jobsv1.KindPolicy{},
			platformasync.ErrDatabaseJobOutcomeInvalid
	}
	return value, policy, nil
}
