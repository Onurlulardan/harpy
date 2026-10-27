package retention

import (
	"context"
	"fmt"
	"time"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/retention"
	"github.com/google/uuid"
)

func (storage *store) Execute(
	ctx context.Context, claim retentionapp.Claim, nextID uuid.UUID,
) (retentionapp.Outcome, error) {
	if storage == nil || storage.database.QueryRow == nil || ctx == nil ||
		ctx.Err() != nil || nextID == uuid.Nil ||
		nextID.Version() != uuid.Version(7) {
		return retentionapp.Outcome{}, retentionapp.ErrClaimInvalid
	}
	var outcome retentionapp.Outcome
	var successor retentionapp.Successor
	var runAt time.Time
	err := storage.database.QueryRow(ctx, executeCycleSQL,
		claim.TenantID, claim.JobID, claim.AttemptID, nextID,
		claim.PayloadReference, claim.PayloadSHA256, retentionapp.BatchLimit,
	).Scan(
		&successor.JobID, &runAt, &successor.PayloadReference,
		&successor.PayloadSHA256, &successor.DedupeKey, &successor.Priority,
		&outcome.Counts.Drafts, &outcome.Counts.Versions,
		&outcome.Counts.Releases, &outcome.Counts.Executions,
		&outcome.Counts.Audits, &outcome.Counts.Capabilities,
		&outcome.LegalHoldBlocked, &outcome.HasRemaining,
		&outcome.CheckpointRevision,
	)
	if err != nil {
		return retentionapp.Outcome{}, fmt.Errorf("execute Analytics retention cycle: %w", err)
	}
	successor.RunAt = runAt.UTC()
	outcome.Successor = successor
	if err = outcome.Validate(claim.TenantID); err != nil {
		return retentionapp.Outcome{}, err
	}
	return outcome, nil
}
