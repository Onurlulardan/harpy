package environmentoperator

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func (coordinator *Coordinator) claim(
	ctx context.Context,
	manifest installationv1.Manifest,
	operationID uuid.UUID,
	revision int64,
) (leaseClaim, error) {
	secret := make([]byte, 32)
	if _, err := io.ReadFull(coordinator.entropy, secret); err != nil {
		return leaseClaim{}, ErrResumable
	}
	digest := sha256.Sum256(secret)
	claim := leaseClaim{attemptID: attemptID(manifest, revision),
		tokenSHA256: digestString(digest)}
	return coordinator.acquire(ctx, operationID, claim, revision)
}

func (coordinator *Coordinator) renew(
	ctx context.Context,
	operationID uuid.UUID,
	claim leaseClaim,
	revision int64,
) (leaseClaim, error) {
	return coordinator.acquire(ctx, operationID, claim, revision)
}

func (coordinator *Coordinator) acquire(
	ctx context.Context,
	operationID uuid.UUID,
	claim leaseClaim,
	revision int64,
) (leaseClaim, error) {
	lease, err := coordinator.onboarding.Claim(ctx, tenantapi.ClaimAttemptRequest{
		OperationID: operationID, AttemptID: claim.attemptID,
		LeaseOwner: coordinator.leaseOwner, LeaseTokenSHA256: claim.tokenSHA256,
		ExpectedRevision: revision, LeaseDuration: maximumLease,
	})
	if err != nil {
		if errors.Is(err, tenantapi.ErrOnboardingLeaseHeld) ||
			errors.Is(err, tenantapi.ErrOnboardingConflict) {
			return leaseClaim{}, ErrResumable
		}
		return leaseClaim{}, ErrResumable
	}
	if lease.OperationID() != operationID || lease.AttemptID() != claim.attemptID {
		return leaseClaim{}, ErrFenced
	}
	claim.revision, claim.leaseUntil = lease.Revision(), lease.LeaseUntil()
	return claim, nil
}

func digestString(value [32]byte) string {
	const alphabet = "0123456789abcdef"
	result := make([]byte, 64)
	for index, item := range value {
		result[index*2] = alphabet[item>>4]
		result[index*2+1] = alphabet[item&15]
	}
	return string(result)
}
