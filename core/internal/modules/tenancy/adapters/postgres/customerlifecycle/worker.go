package customerlifecycle

import (
	"context"
	"time"

	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
	"github.com/google/uuid"
)

func (store *Store) Claim(ctx context.Context, workerID uuid.UUID, now time.Time,
	lease time.Duration,
) (lifecycle.ClaimedOperation, error) {
	if !store.usable(ctx) || workerID == uuid.Nil || now.IsZero() || lease < time.Second {
		return lifecycle.ClaimedOperation{}, lifecycle.ErrInvalid
	}
	var claim lifecycle.ClaimedOperation
	var completedAt *time.Time
	var failureCode *string
	var retentionID, backupID, contributorID *uuid.UUID
	var evidence []byte
	var verifiedAt *time.Time
	err := store.queryRow(ctx, claimSQL, workerID, now, int(lease.Seconds())).Scan(
		&claim.ID, &claim.ControlTenantID, &claim.TargetTenantID, &claim.Kind, &claim.State,
		&claim.Revision, &claim.Replay, &claim.CreatedAt, &claim.UpdatedAt,
		&completedAt, &failureCode, &claim.SchemaName, &retentionID, &backupID,
		&contributorID, &evidence, &verifiedAt,
	)
	if err != nil {
		return lifecycle.ClaimedOperation{}, store.classify(err)
	}
	claim.WorkerID = workerID
	if retentionID != nil {
		claim.Proof.RetentionReceiptID = *retentionID
	}
	if backupID != nil {
		claim.Proof.BackupReceiptID = *backupID
	}
	if contributorID != nil {
		claim.Proof.ContributorReceiptID = *contributorID
	}
	if len(evidence) == 32 {
		copy(claim.Proof.EvidenceSHA256[:], evidence)
	}
	if verifiedAt != nil {
		claim.Proof.VerifiedAt = verifiedAt.UTC()
	}
	claim.CreatedAt, claim.UpdatedAt = claim.CreatedAt.UTC(), claim.UpdatedAt.UTC()
	if completedAt != nil {
		claim.CompletedAt = completedAt.UTC()
	}
	if failureCode != nil {
		claim.FailureCode = *failureCode
	}
	if claim.ID == uuid.Nil || claim.ControlTenantID == uuid.Nil ||
		claim.TargetTenantID == uuid.Nil || claim.Kind == "" || claim.State == "" ||
		claim.Revision < 1 || claim.CreatedAt.IsZero() || claim.UpdatedAt.IsZero() {
		return lifecycle.ClaimedOperation{}, lifecycle.ErrUnavailable
	}
	return claim, nil
}

func (store *Store) CompleteOffboard(ctx context.Context, claim lifecycle.ClaimedOperation,
	now time.Time,
) (lifecycle.Operation, error) {
	if !store.usable(ctx) || claim.ID == uuid.Nil || claim.WorkerID == uuid.Nil {
		return lifecycle.Operation{}, lifecycle.ErrInvalid
	}
	result, _, err := scanOperation(store.queryRow(ctx, completeOffboardSQL,
		claim.ID, claim.WorkerID, now), store.classify)
	return result, err
}

func (store *Store) CompletePurge(ctx context.Context, claim lifecycle.ClaimedOperation,
	receipt lifecycle.ExecutionReceipt, now time.Time,
) (lifecycle.Operation, error) {
	if !store.usable(ctx) || claim.ID == uuid.Nil || claim.WorkerID == uuid.Nil ||
		!receipt.Valid(now) {
		return lifecycle.Operation{}, lifecycle.ErrInvalid
	}
	result, _, err := scanOperation(store.queryRow(ctx, completePurgeSQL,
		claim.ID, claim.WorkerID, receipt.ID, receipt.SHA256[:], now), store.classify)
	return result, err
}

func (store *Store) Fail(ctx context.Context, claim lifecycle.ClaimedOperation,
	code string, now time.Time,
) error {
	if !store.usable(ctx) || claim.ID == uuid.Nil || claim.WorkerID == uuid.Nil || code == "" {
		return lifecycle.ErrInvalid
	}
	var applied bool
	if err := store.queryRow(ctx, failSQL, claim.ID, claim.WorkerID, code, now).Scan(&applied); err != nil {
		return store.classify(err)
	}
	if !applied {
		return lifecycle.ErrStateConflict
	}
	return nil
}
