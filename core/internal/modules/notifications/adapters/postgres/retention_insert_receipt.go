package postgres

import (
	"context"
	"fmt"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/retention"
)

func (store *RetentionStore) insertReceipt(ctx context.Context, cycle retentionapp.Cycle,
	outcome retentionapp.Outcome,
) error {
	claim, counts, next := cycle.Claim, outcome.Counts, outcome.Successor
	digest := receiptDigest(claim.JobID, claim.AttemptID, claim.TenantID,
		claim.PayloadReference, claim.PayloadSHA256, outcome)
	var inserted string
	err := store.database.QueryRow(ctx, insertReceiptSQL,
		claim.JobID, claim.AttemptID, claim.TenantID, claim.PayloadReference,
		claim.PayloadSHA256, next.JobID, next.RunAt, next.DedupeKey,
		counts.ContentRedacted, counts.AttemptsPurged, counts.ReceiptsPurged,
		counts.EffectsPurged, counts.DeliveriesPurged, counts.MessagesPurged,
		counts.HintsPurged, counts.PlatformRecordsPurged, outcome.HasRemaining,
		outcome.CheckpointRevision, digest[:], cycle.CompletedAt,
	).Scan(&inserted)
	if err != nil {
		return fmt.Errorf("insert notification retention receipt: %w", err)
	}
	if inserted != claim.JobID.String() {
		return fmt.Errorf("insert notification retention receipt: identity mismatch")
	}
	return nil
}
