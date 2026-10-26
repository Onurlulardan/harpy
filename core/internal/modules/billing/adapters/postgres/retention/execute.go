package retention

import (
	"context"
	"fmt"
	"time"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/retention"
	"github.com/google/uuid"
)

func (storage *store) Execute(
	ctx context.Context,
	claim retentionapp.Claim,
	nextID uuid.UUID,
) (retentionapp.Outcome, error) {
	authorityID, err := storage.authorityID(ctx, claim.TenantID)
	if err != nil {
		return retentionapp.Outcome{}, fmt.Errorf("prepare billing retention cycle: %w", err)
	}
	if nextID == uuid.Nil || nextID.Version() != uuid.Version(7) {
		return retentionapp.Outcome{}, retentionapp.ErrConfiguration
	}
	var outcome retentionapp.Outcome
	var successor retentionapp.Successor
	var runAt time.Time
	err = storage.database.QueryRow(
		ctx, executeCycleSQL, authorityID, claim.TenantID, claim.JobID,
		claim.AttemptID, nextID, claim.PayloadReference, claim.PayloadSHA256,
	).Scan(
		&successor.JobID, &runAt, &successor.PayloadReference,
		&successor.PayloadSHA256, &successor.DedupeKey, &successor.Priority,
		&outcome.RawBodiesRedacted, &outcome.RawRecordsPurged,
		&outcome.EntitlementPurged, &outcome.FinancialPurged,
		&outcome.LegalHoldBlocked, &outcome.CheckpointRevision,
	)
	if err != nil {
		return retentionapp.Outcome{}, fmt.Errorf("execute billing retention cycle: %w", err)
	}
	successor.RunAt = runAt.UTC()
	outcome.Successor = successor
	return outcome, nil
}
