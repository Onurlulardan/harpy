package postgres

import (
	"context"
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func (store *boundDatabaseJobStore) enqueueSuccessor(
	ctx context.Context,
	claim JobClaim,
	policy jobsv1.KindPolicy,
	successor platformasync.DatabaseJobSuccessorInput,
) error {
	if successor.JobID == claim.JobID() || !successor.RunAt.After(claim.StartedAt()) {
		return platformasync.ErrDatabaseJobOutcomeInvalid
	}
	tenantID := claim.tenantID.String()
	value, err := jobsv1.NewEnqueue(jobsv1.Enqueue{
		JobID: successor.JobID.String(), Kind: policy.Kind,
		PayloadRevision: policy.PayloadRevision, TenantID: &tenantID,
		Authority: jobsv1.Authority{
			ScopeKind: jobsv1.ScopeTenant, ScopeID: tenantID,
			Domain:            policy.OwnerDomain,
			CutoverGeneration: claim.authority.generation,
			WriterOrigin:      jobsv1.WriterV2, GroupID: claim.authority.groupID,
			GroupGeneration: claim.authority.groupGeneration,
		},
		PayloadReference: successor.PayloadReference,
		PayloadSHA256:    successor.PayloadSHA256,
		RunAt:            successor.RunAt.Format(time.RFC3339Nano),
		Priority:         successor.Priority, DedupeKey: successor.DedupeKey,
		MaximumAttempts: policy.MaximumAttempts,
	}, mustSingleJobCatalog(policy))
	if err != nil {
		return fmt.Errorf("validate database job successor: %w", err)
	}
	result, err := store.database.Execute(
		ctx, enqueueJobSQL, successor.JobID, claim.tenantID, "tenant",
		claim.tenantID, claim.authority.groupID, claim.authority.groupGeneration,
		claim.authority.domain, claim.authority.generation, "v2", policy.Kind,
		policy.PayloadRevision, successor.PayloadReference, successor.PayloadSHA256,
		successor.DedupeKey, string(policy.EffectClass), nil,
		successor.Priority, successor.RunAt, policy.MaximumAttempts, nil,
	)
	if err != nil {
		return fmt.Errorf("insert database job successor: %w", err)
	}
	if result.RowsAffected() == 1 {
		return nil
	}
	return store.assertExactSuccessor(
		ctx, claim, value, successor.JobID, successor.RunAt,
	)
}

func (store *boundDatabaseJobStore) assertExactSuccessor(
	ctx context.Context,
	claim JobClaim,
	value jobsv1.Enqueue,
	jobID uuid.UUID,
	runAt time.Time,
) error {
	var existingID string
	var exact bool
	err := store.database.QueryRow(
		ctx, findJobEnqueueSQL, jobID, claim.tenantID,
		claim.authority.groupID, claim.authority.domain, value.Kind,
		value.DedupeKey, claim.authority.groupGeneration,
		claim.authority.generation, value.PayloadRevision, value.PayloadSHA256,
		string(jobsv1.EffectDatabaseTransactional), nil, value.Priority,
		value.MaximumAttempts, nil, runAt, value.PayloadReference,
	).Scan(&existingID, &exact)
	if err != nil {
		return fmt.Errorf("read database job successor: %w", err)
	}
	if !exact || existingID != jobID.String() {
		return ErrIdempotencyConflict
	}
	return nil
}

func mustSingleJobCatalog(policy jobsv1.KindPolicy) jobsv1.Catalog {
	catalog, err := jobsv1.NewCatalog([]jobsv1.KindPolicy{policy})
	if err != nil {
		return jobsv1.Catalog{}
	}
	return catalog
}
