package systemtransfer

import (
	"context"

	"github.com/google/uuid"
)

func ClaimSourceCleanup(ctx context.Context,
	store SourceCleanupStore,
	claim SourceCleanupClaim,
) (SourceCleanupScan, error) {
	if ctx == nil || ctx.Err() != nil || nilDependency(store) ||
		claim.LeaseOwnerID.Version() != uuid.Version(7) ||
		!controlTime(claim.ObservedAt) {
		return SourceCleanupScan{}, ErrSourceCleanupInvalid
	}
	scan, err := store.AcquireSourceCleanup(ctx, SourceCleanupClaim{
		AfterTenantID: claim.AfterTenantID, LeaseOwnerID: claim.LeaseOwnerID,
		ObservedAt: claim.ObservedAt.UTC(),
	})
	if err != nil {
		return SourceCleanupScan{}, err
	}
	if scan.Validate() != nil {
		return SourceCleanupScan{}, ErrSourceCleanupInvalid
	}
	return scan, nil
}
