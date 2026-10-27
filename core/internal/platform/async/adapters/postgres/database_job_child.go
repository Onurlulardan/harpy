package postgres

import (
	"context"
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func (store *boundDatabaseJobStore) enqueueChild(
	ctx context.Context,
	claim JobClaim,
	child platformasync.DatabaseJobChildInput,
) error {
	value, policy, err := newDatabaseJobChildEnqueue(claim, child, store.jobs)
	if err != nil {
		return err
	}
	result, err := store.database.Execute(
		ctx, enqueueJobSQL, child.JobID, claim.tenantID, "tenant",
		claim.tenantID, claim.authority.groupID, claim.authority.groupGeneration,
		claim.authority.domain, claim.authority.generation, "v2", policy.Kind,
		policy.PayloadRevision, child.PayloadReference, child.PayloadSHA256,
		child.DedupeKey, string(policy.EffectClass), value.EffectIdempotencyKey,
		child.Priority, child.RunAt, policy.MaximumAttempts, nil,
	)
	if err != nil {
		return fmt.Errorf("insert database job child: %w", err)
	}
	if result.RowsAffected() == 1 {
		return nil
	}
	return store.assertExactChild(ctx, claim, value, policy, child)
}

func (store *boundDatabaseJobStore) assertExactChild(
	ctx context.Context,
	claim JobClaim,
	value jobsv1.Enqueue,
	policy jobsv1.KindPolicy,
	child platformasync.DatabaseJobChildInput,
) error {
	var existingID string
	var exact bool
	err := store.database.QueryRow(
		ctx, findJobEnqueueSQL, child.JobID, claim.tenantID,
		claim.authority.groupID, claim.authority.domain, value.Kind,
		value.DedupeKey, claim.authority.groupGeneration,
		claim.authority.generation, value.PayloadRevision, value.PayloadSHA256,
		string(policy.EffectClass), value.EffectIdempotencyKey, value.Priority,
		value.MaximumAttempts, nil, child.RunAt, value.PayloadReference,
	).Scan(&existingID, &exact)
	if err != nil {
		return fmt.Errorf("read database job child: %w", err)
	}
	if !exact || existingID != child.JobID.String() {
		return ErrIdempotencyConflict
	}
	return nil
}
