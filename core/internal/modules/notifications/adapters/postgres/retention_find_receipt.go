package postgres

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/retention"
	"github.com/google/uuid"
)

type storedReceipt struct {
	attemptID, tenantID uuid.UUID
	reference, digest   string
	outcome             retentionapp.Outcome
	receiptDigest       []byte
	completedAt         time.Time
}

func (store *RetentionStore) findReceipt(ctx context.Context,
	claim retentionapp.Claim,
) (retentionapp.Outcome, bool, error) {
	var raw storedReceipt
	var attempt, tenant, successorID string
	scanErr := store.database.QueryRow(ctx, findReceiptSQL, claim.JobID).Scan(
		&attempt, &tenant, &raw.reference, &raw.digest, &successorID,
		&raw.outcome.Successor.RunAt, &raw.outcome.Successor.DedupeKey,
		&raw.outcome.Counts.ContentRedacted, &raw.outcome.Counts.AttemptsPurged,
		&raw.outcome.Counts.ReceiptsPurged, &raw.outcome.Counts.EffectsPurged,
		&raw.outcome.Counts.DeliveriesPurged, &raw.outcome.Counts.MessagesPurged,
		&raw.outcome.Counts.HintsPurged, &raw.outcome.Counts.PlatformRecordsPurged,
		&raw.outcome.HasRemaining, &raw.outcome.CheckpointRevision,
		&raw.receiptDigest, &raw.completedAt)
	if scanErr != nil && store.database.RowNotFound(scanErr) {
		return retentionapp.Outcome{}, false, nil
	}
	if scanErr != nil {
		return retentionapp.Outcome{}, false, fmt.Errorf("find retention receipt: %w", scanErr)
	}
	var parseErr error
	raw.attemptID, parseErr = uuid.Parse(attempt)
	if parseErr == nil {
		raw.tenantID, parseErr = uuid.Parse(tenant)
	}
	if parseErr == nil {
		raw.outcome.Successor.JobID, parseErr = uuid.Parse(successorID)
	}
	raw.outcome.Successor.PayloadReference = raw.reference
	raw.outcome.Successor.PayloadSHA256 = raw.digest
	raw.outcome.Successor.RunAt = raw.outcome.Successor.RunAt.UTC()
	expected := receiptDigest(claim.JobID, raw.attemptID, raw.tenantID,
		raw.reference, raw.digest, raw.outcome)
	if parseErr != nil || raw.tenantID != claim.TenantID || raw.reference != claim.PayloadReference ||
		raw.digest != claim.PayloadSHA256 || len(raw.receiptDigest) != len(expected) ||
		subtle.ConstantTimeCompare(raw.receiptDigest, expected[:]) != 1 ||
		raw.outcome.Validate(claim.TenantID) != nil {
		return retentionapp.Outcome{}, false, errors.Join(
			retentionapp.ErrReceiptConflict, parseErr)
	}
	return raw.outcome, true, nil
}
