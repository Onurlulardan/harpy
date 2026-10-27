package recoveryexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *Service) ClaimNext(
	ctx context.Context,
	attemptID string,
	leaseToken string,
) (ports.RecoveryExecutionClaim, bool, error) {
	if service == nil || ctx == nil || ctx.Err() != nil ||
		!validUUIDv7(attemptID) || !validUUIDv7(leaseToken) {
		return ports.RecoveryExecutionClaim{}, false, ErrUnavailable
	}
	claim, found, err := service.store.ClaimNextRecoverySet(
		ctx, attemptID, leaseToken,
	)
	if err != nil {
		return ports.RecoveryExecutionClaim{}, false, ErrUnavailable
	}
	if !found {
		return ports.RecoveryExecutionClaim{}, false, nil
	}
	if !validClaim(claim) {
		return ports.RecoveryExecutionClaim{}, false, ErrUnavailable
	}
	return claim, true, nil
}

func validClaim(claim ports.RecoveryExecutionClaim) bool {
	components, err := recoveryset.ParseComponentInventory(
		claim.ComponentInventory, claim.RecoveryState == "verifying",
	)
	return err == nil &&
		validUUIDv7(claim.AttemptID) && validUUIDv7(claim.LeaseToken) &&
		claim.AttemptGeneration > 0 &&
		validRecoveryState(claim.RecoveryState) &&
		validUUID(claim.RecoverySetID) && validUUID(claim.OperationID) &&
		validUUID(claim.EnvironmentID) && claim.Sequence > 0 &&
		validUUID(claim.PrimaryBindingID) && validDigest(claim.TenantSetSHA256) &&
		validUUID(claim.OwnerBindingID) && claim.OwnerBindingGeneration > 0 &&
		validUUID(claim.DatabaseLineageID) && validUUID(claim.DatabaseSystemID) &&
		claim.SigningGeneration > 0 &&
		validDigest(claim.RecoveryTrustBundleSHA256) &&
		len(components) <= 5 && !claim.DatabaseNow.IsZero()
}

func validRecoveryState(value string) bool {
	return value == "holding" || value == "snapshotting" ||
		value == "copying" || value == "verifying"
}
