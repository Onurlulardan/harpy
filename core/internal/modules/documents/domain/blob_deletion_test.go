package domain

import (
	"crypto/sha256"
	"errors"
	"testing"
	"time"
)

func TestBlobDeletionFencesStaleWorkerAndRequiresAbsenceEvidence(t *testing.T) {
	now := time.Date(2026, 7, 18, 10, 0, 0, 0, time.UTC)
	owner := mustID(t, "018f0000-0000-7000-8000-000000000401")
	deletion := BlobDeletion{ID: mustID(t, "018f0000-0000-7000-8000-000000000402"),
		TenantID: mustID(t, "11111111-1111-4111-8111-111111111111"),
		TargetID: mustID(t, "018f0000-0000-7000-8000-000000000403"),
		State:    BlobDeletionPending}
	claimed, err := deletion.Claim(owner, now, time.Minute)
	if err != nil || claimed.LeaseGeneration != 1 {
		t.Fatalf("claimed = %+v, %v", claimed, err)
	}
	if _, err = claimed.Reconcile(owner, 2); !errors.Is(err, ErrBlobDeletionTransition) {
		t.Fatalf("stale generation error = %v", err)
	}
	reconciling, _ := claimed.Reconcile(owner, 1)
	confirmed, err := reconciling.Confirm(sha256.Sum256([]byte("delete")),
		sha256.Sum256([]byte("absent")),
		mustID(t, "018f0000-0000-7000-8000-000000000404"), now)
	if err != nil || confirmed.State != BlobDeletionConfirmed {
		t.Fatalf("confirmed = %+v, %v", confirmed, err)
	}
}
