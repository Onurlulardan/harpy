package restoreexecution

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *Service) Claim(
	ctx context.Context,
	identity ports.RestoreExecutionIdentity,
	planSHA256 string,
) (ports.RestoreExecutionClaim, bool, error) {
	if service == nil || service.store == nil || ctx == nil ||
		ctx.Err() != nil || !identity.Valid() ||
		!digestPattern.MatchString(planSHA256) {
		return ports.RestoreExecutionClaim{}, false, ErrClaim
	}
	claim, found, err := service.store.ClaimRestoreExecution(
		ctx, identity, planSHA256,
	)
	if err != nil {
		return ports.RestoreExecutionClaim{}, false, ErrClaim
	}
	if !found {
		return ports.RestoreExecutionClaim{}, false, nil
	}
	if !validClaim(claim, identity, planSHA256) {
		return ports.RestoreExecutionClaim{}, false, ErrClaim
	}
	return claim, true, nil
}

func validClaim(
	claim ports.RestoreExecutionClaim,
	identity ports.RestoreExecutionIdentity,
	planSHA256 string,
) bool {
	planDigest := sha256.Sum256(claim.PlanBytes)
	manifestDigest := sha256.Sum256(claim.ManifestBytes)
	return claim.RestoreExecutionIdentity == identity &&
		claim.AttemptGeneration > 0 &&
		validState(claim.RestoreState) &&
		canonicalUUID.MatchString(claim.RecoverySetID) &&
		canonicalUUID.MatchString(claim.EnvironmentID) &&
		claim.PlanSHA256 == planSHA256 &&
		claim.PlanSHA256 == hex.EncodeToString(planDigest[:]) &&
		claim.ManifestSHA256 == hex.EncodeToString(manifestDigest[:]) &&
		canonicalUUID.MatchString(claim.PreviousDatabaseSystemID) &&
		validSafety(claim) && !claim.DatabaseNow.IsZero() &&
		validComponentInventory(claim.ComponentInventory) &&
		validResumeEvidence(claim)
}

func validState(value string) bool {
	switch value {
	case "plan_ready", "draining", "restoring", "verifying", "switched":
		return true
	default:
		return false
	}
}

func validSafety(claim ports.RestoreExecutionClaim) bool {
	switch claim.TargetDisposition {
	case "existing_readable":
		return canonicalUUID.MatchString(claim.SafetyRecoverySetID) &&
			digestPattern.MatchString(claim.SafetyRecoverySetSHA256) &&
			claim.WaiverReason == "" && claim.ProbeSHA256 == "" &&
			!claim.DestructiveConfirmed
	case "empty":
		return claim.SafetyRecoverySetID == "" &&
			claim.SafetyRecoverySetSHA256 == "" &&
			claim.WaiverReason == "empty_no_database_or_policy" &&
			digestPattern.MatchString(claim.ProbeSHA256) &&
			!claim.DestructiveConfirmed
	case "unreadable":
		return claim.SafetyRecoverySetID == "" &&
			claim.SafetyRecoverySetSHA256 == "" &&
			claim.WaiverReason == "unreadable_integrity_probe" &&
			digestPattern.MatchString(claim.ProbeSHA256) &&
			claim.DestructiveConfirmed
	default:
		return false
	}
}

func validResumeEvidence(claim ports.RestoreExecutionClaim) bool {
	emptyApply := claim.ControlDatabaseLocatorSHA256 == "" &&
		claim.ScratchDatabaseLocatorSHA256 == "" &&
		claim.DrainAckSHA256 == ""
	emptyVerify := claim.ArchiveReadbackSHA256 == "" &&
		claim.NewDatabaseSystemID == ""
	emptySwitch := claim.SwitchAckSHA256 == "" &&
		claim.ObjectMerkleSHA256 == "" &&
		claim.LocatorRemapSHA256 == "" &&
		claim.SafetyEvidenceSHA256 == ""
	if claim.RestoreState == "plan_ready" || claim.RestoreState == "draining" {
		return emptyApply && emptyVerify && emptySwitch
	}
	validApply := digestPattern.MatchString(claim.ControlDatabaseLocatorSHA256) &&
		digestPattern.MatchString(claim.ScratchDatabaseLocatorSHA256) &&
		claim.ControlDatabaseLocatorSHA256 != claim.ScratchDatabaseLocatorSHA256 &&
		digestPattern.MatchString(claim.DrainAckSHA256)
	if claim.RestoreState == "restoring" {
		return validApply && emptyVerify && emptySwitch
	}
	validVerify := digestPattern.MatchString(claim.ArchiveReadbackSHA256) &&
		canonicalUUID.MatchString(claim.NewDatabaseSystemID) &&
		claim.NewDatabaseSystemID != claim.PreviousDatabaseSystemID
	if claim.RestoreState == "verifying" {
		return validApply && validVerify && emptySwitch
	}
	return validApply && validVerify &&
		digestPattern.MatchString(claim.SwitchAckSHA256) &&
		digestPattern.MatchString(claim.ObjectMerkleSHA256) &&
		digestPattern.MatchString(claim.LocatorRemapSHA256) &&
		digestPattern.MatchString(claim.SafetyEvidenceSHA256)
}
