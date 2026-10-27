package postgres

import (
	"context"
	"time"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/retention"
	"github.com/google/uuid"
)

func (store *RetentionStore) ApplyCycle(ctx context.Context,
	cycle retentionapp.Cycle,
) (retentionapp.Outcome, error) {
	if store == nil || ctx == nil || ctx.Err() != nil || cycle.Claim.Check() != nil ||
		cycle.CompletedAt.IsZero() || cycle.CompletedAt.Location() != time.UTC ||
		cycle.CompletedAt.Nanosecond()%int(time.Microsecond) != 0 ||
		cycle.SuccessorJobID == uuid.Nil ||
		cycle.SuccessorJobID.Version() != uuid.Version(7) {
		return retentionapp.Outcome{}, retentionapp.ErrClaimInvalid
	}
	revision, err := store.lock(ctx)
	if err != nil {
		return retentionapp.Outcome{}, err
	}
	if replay, found, findErr := store.findReceipt(ctx, cycle.Claim); findErr != nil {
		return retentionapp.Outcome{}, findErr
	} else if found {
		return replay, nil
	}
	counts, err := store.pruneRecords(ctx, cycle)
	if err != nil {
		return retentionapp.Outcome{}, err
	}
	remaining, err := store.hasRemaining(ctx, cycle.Claim.TenantID, cycle.CompletedAt)
	if err != nil {
		return retentionapp.Outcome{}, err
	}
	outcome := retentionapp.Outcome{Counts: counts, HasRemaining: remaining,
		CheckpointRevision: revision + 1,
		Successor: successor(cycle.CompletedAt, cycle.Claim.TenantID,
			cycle.SuccessorJobID, remaining)}
	if err = outcome.Validate(cycle.Claim.TenantID); err != nil {
		return retentionapp.Outcome{}, err
	}
	if err = store.insertReceipt(ctx, cycle, outcome); err != nil {
		return retentionapp.Outcome{}, err
	}
	if err = store.persistCheckpoint(ctx, cycle, revision); err != nil {
		return retentionapp.Outcome{}, err
	}
	return outcome, nil
}
