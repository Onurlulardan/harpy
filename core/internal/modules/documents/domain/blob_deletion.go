package domain

import (
	"errors"
	"time"
)

var ErrBlobDeletionTransition = errors.New("documents blob deletion transition is invalid")

type BlobDeletionState string

const (
	BlobDeletionPending     BlobDeletionState = "pending"
	BlobDeletionClaimed     BlobDeletionState = "claimed"
	BlobDeletionReconciling BlobDeletionState = "reconciling"
	BlobDeletionConfirmed   BlobDeletionState = "confirmed"
)

type BlobDeletion struct {
	ID, TenantID, TargetID, LeaseOwnerID, BillingReceiptID ID
	TargetKind, ObjectKey, ObjectVersion                   string
	ObjectGeneration, LeaseGeneration                      uint64
	State                                                  BlobDeletionState
	LeaseExpiresAt, ConfirmedAt                            time.Time
	DeleteSHA256, AbsenceSHA256                            [32]byte
}

func (deletion BlobDeletion) Claim(owner ID, now time.Time, lease time.Duration) (BlobDeletion, error) {
	claimable := deletion.State == BlobDeletionPending ||
		deletion.State == BlobDeletionClaimed && !now.Before(deletion.LeaseExpiresAt)
	if !claimable || !owner.UUIDv7() || !validInstant(now) || lease <= 0 || lease > time.Minute {
		return BlobDeletion{}, ErrBlobDeletionTransition
	}
	deletion.State, deletion.LeaseOwnerID = BlobDeletionClaimed, owner
	deletion.LeaseGeneration++
	deletion.LeaseExpiresAt = now.UTC().Add(lease)
	return deletion, nil
}

func (deletion BlobDeletion) Reconcile(owner ID, generation uint64) (BlobDeletion, error) {
	if deletion.State != BlobDeletionClaimed || deletion.LeaseOwnerID != owner ||
		deletion.LeaseGeneration != generation {
		return BlobDeletion{}, ErrBlobDeletionTransition
	}
	deletion.State = BlobDeletionReconciling
	return deletion, nil
}

func (deletion BlobDeletion) Confirm(deleteDigest, absenceDigest [32]byte,
	billingReceipt ID, now time.Time,
) (BlobDeletion, error) {
	if deletion.State != BlobDeletionReconciling || deleteDigest == [32]byte{} ||
		absenceDigest == [32]byte{} || !billingReceipt.UUIDv7() || !validInstant(now) {
		return BlobDeletion{}, ErrBlobDeletionTransition
	}
	deletion.State, deletion.ConfirmedAt = BlobDeletionConfirmed, now.UTC()
	deletion.DeleteSHA256, deletion.AbsenceSHA256 = deleteDigest, absenceDigest
	deletion.BillingReceiptID = billingReceipt
	return deletion, nil
}
