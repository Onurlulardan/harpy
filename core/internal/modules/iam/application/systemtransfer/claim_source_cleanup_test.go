package systemtransfer

import (
	"context"
	"errors"
	"testing"
)

func TestClaimSourceCleanupReturnsValidatedTenantTarget(t *testing.T) {
	target, _, now := importCleanupFixture()
	store := &importCleanupStoreStub{scans: []SourceCleanupScan{{
		ScannedTenantID: target.TenantID, Target: target, Found: true,
	}}}
	scan, err := ClaimSourceCleanup(context.Background(), store, SourceCleanupClaim{
		LeaseOwnerID: target.LeaseOwnerID, ObservedAt: now,
	})
	if err != nil || !scan.Found || scan.Target != target ||
		store.claim.ObservedAt.Location() != now.Location() {
		t.Fatalf("ClaimSourceCleanup() = %#v, %v", scan, err)
	}
}

func TestClaimSourceCleanupRejectsInvalidOwner(t *testing.T) {
	_, err := ClaimSourceCleanup(context.Background(), &importCleanupStoreStub{},
		SourceCleanupClaim{})
	if !errors.Is(err, ErrSourceCleanupInvalid) {
		t.Fatalf("error = %v", err)
	}
}
